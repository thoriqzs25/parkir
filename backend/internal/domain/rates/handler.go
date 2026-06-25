package rates

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thoriqzs/PARKIR/backend/internal/errors"
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

type CreateRateRequest struct {
	VehicleType          string  `json:"vehicle_type" binding:"required,oneof=CAR MOTO TRUCK"`
	FirstHourRate        float64 `json:"first_hour_rate" binding:"required,gte=0"`
	SubsequentHourlyRate float64 `json:"subsequent_hourly_rate" binding:"required,gte=0"`
	DailyFlatRate        float64 `json:"daily_flat_rate" binding:"required,gte=0"`
	EffectiveFrom        string  `json:"effective_from" binding:"required"`
	EffectiveUntil       *string `json:"effective_until,omitempty"`
}

type UpdateRateRequest struct {
	FirstHourRate        *float64 `json:"first_hour_rate,omitempty"`
	SubsequentHourlyRate *float64 `json:"subsequent_hourly_rate,omitempty"`
	DailyFlatRate        *float64 `json:"daily_flat_rate,omitempty"`
	EffectiveUntil       *string  `json:"effective_until,omitempty"`
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func parseOptionalDate(s *string) (*time.Time, error) {
	if s == nil {
		return nil, nil
	}
	t, err := parseDate(*s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (h *Handler) RegisterRoutes(locations *gin.RouterGroup, rates *gin.RouterGroup) {
	locRates := locations.Group("/:id/rates")
	locRates.Use(middleware.RequirePermission("rates:view"))
	{
		locRates.GET("", h.List)
	}

	locRatesWithManage := locations.Group("/:id/rates")
	locRatesWithManage.Use(middleware.RequirePermission("rates:create"))
	{
		locRatesWithManage.POST("", h.Create)
	}

	singleRate := rates.Group("/:id")
	singleRate.Use(middleware.RequirePermission("rates:edit"))
	{
		singleRate.PATCH("", h.Update)
	}
}

func (h *Handler) Create(c *gin.Context) {
	locationID := c.Param("id")

	var req CreateRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	effectiveFrom, err := parseDate(req.EffectiveFrom)
	if err != nil {
		response.BadRequest(c, "INVALID_DATE", "effective_from must be YYYY-MM-DD")
		return
	}

	effectiveUntil, err := parseOptionalDate(req.EffectiveUntil)
	if err != nil {
		response.BadRequest(c, "INVALID_DATE", "effective_until must be YYYY-MM-DD")
		return
	}

	rate, err := h.store.CreateRate(c.Request.Context(), store.CreateRateInput{
		LocationID:           locationID,
		VehicleType:          req.VehicleType,
		FirstHourRate:        req.FirstHourRate,
		SubsequentHourlyRate: req.SubsequentHourlyRate,
		DailyFlatRate:        req.DailyFlatRate,
		EffectiveFrom:        effectiveFrom,
		EffectiveUntil:       effectiveUntil,
		CreatedBy:            middleware.GetUserID(c),
	})
	if err != nil {
		_ = c.Error(err)
		if err == errors.ErrRateOverlap || strings.Contains(err.Error(), "Overlapping rate effective dates") {
			response.Conflict(c, "RATE_OVERLAP", "rate overlaps with existing rate for this vehicle type")
			return
		}
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "rate.created", rate.ID, &locationID, gin.H{"vehicle_type": req.VehicleType, "effective_from": req.EffectiveFrom})
	response.Created(c, rate)
}

func (h *Handler) List(c *gin.Context) {
	locationID := c.Param("id")

	rates, err := h.store.ListRatesByLocation(c.Request.Context(), locationID)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.OK(c, rates)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	effectiveUntil, err := parseOptionalDate(req.EffectiveUntil)
	if err != nil {
		response.BadRequest(c, "INVALID_DATE", "effective_until must be YYYY-MM-DD")
		return
	}

	rate, err := h.store.UpdateRate(c.Request.Context(), id, store.UpdateRateInput{
		FirstHourRate:        req.FirstHourRate,
		SubsequentHourlyRate: req.SubsequentHourlyRate,
		DailyFlatRate:        req.DailyFlatRate,
		EffectiveUntil:       effectiveUntil,
	})
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "rate")
			return
		}
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "rate.updated", id, nil, gin.H{"effective_until": req.EffectiveUntil})
	response.OK(c, rate)
}

func (h *Handler) logAudit(c *gin.Context, action, entityID string, locationID *string, metadata map[string]interface{}) {
	actorID := middleware.GetUserID(c)
	roleName := middleware.GetRoleName(c)
	ip := c.ClientIP()
	_ = h.store.CreateAuditLog(c.Request.Context(), store.AuditLogEntry{
		Action:     action,
		ActorID:    &actorID,
		ActorRole:  &roleName,
		EntityType: "rate",
		EntityID:   entityID,
		LocationID: locationID,
		IPAddress:  &ip,
		Metadata:   metadata,
	})
}
