package sessions

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thoriqzs/PARKIR/backend/internal/errors"
	"github.com/thoriqzs/PARKIR/backend/internal/fees"
	"github.com/thoriqzs/PARKIR/backend/internal/middleware"
	"github.com/thoriqzs/PARKIR/backend/internal/response"
	"github.com/thoriqzs/PARKIR/backend/internal/store"
)

type Handler struct {
	store *store.Store
}

func NewHandler(store *store.Store) *Handler {
	return &Handler{store: store}
}

type CheckInRequest struct {
	LocationID  string `json:"location_id" binding:"required,uuid"`
	Plate       string `json:"plate" binding:"required"`
	CityCode    string `json:"city_code"`
	VehicleType string `json:"vehicle_type" binding:"required,oneof=CAR MOTO TRUCK"`
}

type CheckOutRequest struct {
	FeeAmount *float64 `json:"fee_amount,omitempty"`
}

type SessionResponse struct {
	Session        store.Session `json:"session"`
	DuplicatePlate bool          `json:"duplicate_plate_warning"`
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	sessions := r.Group("/sessions")
	sessions.Use(middleware.RequirePermission("sessions:view"))
	{
		sessions.GET("", h.List)
		sessions.GET("/:id", h.Get)
	}

	sessionsWithCreate := r.Group("/sessions")
	sessionsWithCreate.Use(middleware.RequirePermission("sessions:create"))
	{
		sessionsWithCreate.POST("/check-in", h.CheckIn)
	}

	sessionsWithClose := r.Group("/sessions")
	sessionsWithClose.Use(middleware.RequirePermission("sessions:close"))
	{
		sessionsWithClose.POST("/:id/check-out", h.CheckOut)
	}
}

func normalizePlate(plate string) string {
	return strings.ToUpper(strings.TrimSpace(plate))
}

func (h *Handler) CheckIn(c *gin.Context) {
	var req CheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	operatorID := middleware.GetUserID(c)
	ctx := c.Request.Context()
	checkInAt := time.Now()

	// Auto-detect shift based on check-in time
	shiftConfig, err := h.store.GetShiftConfigByTimeWithFallback(ctx, req.LocationID, checkInAt)
	if err != nil {
		if err == errors.ErrNotFound {
			response.BadRequest(c, "NO_SHIFT_CONFIG", "no shift configuration found for this location and time")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	// Determine shift date (today, unless overnight shift and current time is before end_time)
	shiftDate := time.Date(checkInAt.Year(), checkInAt.Month(), checkInAt.Day(), 0, 0, 0, 0, checkInAt.Location())
	if shiftConfig.IsOvernight {
		endTime, _ := time.Parse("15:04:05", shiftConfig.EndTime)
		if checkInAt.Hour() < endTime.Hour() || (checkInAt.Hour() == endTime.Hour() && checkInAt.Minute() < endTime.Minute()) {
			// Current time is in the "next day" part of overnight shift
			shiftDate = shiftDate.AddDate(0, 0, -1)
		}
	}

	// Get or create shift instance
	shift, err := h.store.GetOrCreateShift(ctx, req.LocationID, shiftConfig.ShiftNumber, shiftDate)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	plate := normalizePlate(req.Plate)

	duplicate := false
	_, dupErr := h.store.FindActiveSessionByPlate(ctx, req.LocationID, plate)
	if dupErr == nil {
		duplicate = true
	}

	session, err := h.store.CreateSession(ctx, store.CreateSessionInput{
		LocationID:  req.LocationID,
		OperatorID:  operatorID,
		ShiftID:     shift.ID,
		Plate:       plate,
		CityCode:    strings.ToUpper(strings.TrimSpace(req.CityCode)),
		VehicleType: req.VehicleType,
	})
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "session.check_in", session.ID, &req.LocationID, gin.H{
		"plate":         plate,
		"vehicle_type":  req.VehicleType,
		"shift_id":      shift.ID,
		"shift_number":  shift.ShiftNumber,
		"shift_date":    shift.ShiftDate,
	})

	response.Created(c, SessionResponse{Session: *session, DuplicatePlate: duplicate})
}

func (h *Handler) CheckOut(c *gin.Context) {
	id := c.Param("id")

	var req CheckOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	session, err := h.store.GetSessionByID(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "session")
			return
		}
		response.InternalServerError(c)
		return
	}

	if session.State != "ACTIVE" {
		response.BadRequest(c, "INVALID_STATE", "session is not active")
		return
	}

	checkOutAt := time.Now().UTC()

	var fee float64
	var durationHours int
	var rateSnapshot map[string]interface{}

	if req.FeeAmount != nil {
		fee = *req.FeeAmount
		durationHours = calculateDurationHours(session.CheckInAt, checkOutAt)
		rateSnapshot = map[string]interface{}{
			"manual_override": true,
			"fee_amount":      fee,
		}
	} else {
		rate, err := h.store.GetActiveRate(c.Request.Context(), session.LocationID, session.VehicleType, session.CheckInAt)
		if err != nil {
			if err == errors.ErrNotFound {
				response.BadRequest(c, "NO_RATE", "no active rate for this vehicle type; provide fee_amount override")
				return
			}
			response.InternalServerError(c)
			return
		}

		fee, durationHours = fees.Calculate(fees.CalculationInput{
			FirstHourRate:        rate.FirstHourRate,
			SubsequentHourlyRate: rate.SubsequentHourlyRate,
			DailyFlatRate:        rate.DailyFlatRate,
			CheckInAt:            session.CheckInAt,
			CheckOutAt:           checkOutAt,
		})

		rateSnapshot = map[string]interface{}{
			"rate_id":                rate.ID,
			"first_hour_rate":        rate.FirstHourRate,
			"subsequent_hourly_rate": rate.SubsequentHourlyRate,
			"daily_flat_rate":        rate.DailyFlatRate,
			"effective_from":         rate.EffectiveFrom,
		}
	}

	session, err = h.store.UpdateSessionToPendingPayment(c.Request.Context(), id, store.CheckOutSessionInput{
		CheckOutAt:   checkOutAt,
		FeeAmount:    &fee,
		RateSnapshot: rateSnapshot,
	})
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "session")
			return
		}
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "session.check_out", session.ID, &session.LocationID, gin.H{
		"fee_amount":     fee,
		"duration_hours": durationHours,
	})

	response.OK(c, session)
}

func calculateDurationHours(checkInAt, checkOutAt time.Time) int {
	diff := checkOutAt.Sub(checkInAt)
	hours := int(diff.Hours())
	if diff.Hours() > float64(hours) {
		hours++
	}
	if hours < 1 {
		hours = 1
	}
	return hours
}

func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filters := store.ListSessionsFilters{
		LocationID: c.Query("location_id"),
		State:      c.Query("state"),
		Plate:      c.Query("plate"),
		OperatorID: c.Query("operator_id"),
	}

	sessions, total, err := h.store.ListSessions(c.Request.Context(), filters, limit, offset)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.OK(c, gin.H{
		"items": sessions,
		"meta":  response.Meta{Limit: limit, Offset: offset, Total: total},
	})
}

func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")
	session, err := h.store.GetSessionByID(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "session")
			return
		}
		response.InternalServerError(c)
		return
	}

	include := c.Query("include")
	if include == "transaction" {
		tx, err := h.store.GetTransactionBySessionID(c.Request.Context(), session.ID)
		if err != nil && err != errors.ErrNotFound {
			response.InternalServerError(c)
			return
		}
		response.OK(c, gin.H{
			"session":     session,
			"transaction": tx,
		})
		return
	}

	response.OK(c, session)
}

func (h *Handler) logAudit(c *gin.Context, action, entityID string, locationID *string, metadata map[string]interface{}) {
	actorID := middleware.GetUserID(c)
	roleName := middleware.GetRoleName(c)
	ip := c.ClientIP()
	_ = h.store.CreateAuditLog(c.Request.Context(), store.AuditLogEntry{
		Action:     action,
		ActorID:    &actorID,
		ActorRole:  &roleName,
		EntityType: "session",
		EntityID:   entityID,
		LocationID: locationID,
		IPAddress:  &ip,
		Metadata:   metadata,
	})
}
