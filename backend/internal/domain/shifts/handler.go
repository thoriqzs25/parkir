package shifts

import (
	"strconv"
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

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// Location-based shift routes
	locationShifts := r.Group("/locations/:id/shifts")
	locationShifts.Use(middleware.RequirePermission("shifts:view"))
	{
		locationShifts.GET("", h.ListByLocation)
		locationShifts.GET("/current", h.GetCurrent)
	}

	// Individual shift routes
	shifts := r.Group("/shifts")
	shifts.Use(middleware.RequirePermission("shifts:view"))
	{
		shifts.GET("/:id", h.Get)
		shifts.GET("/:id/summary", h.GetSummary)
	}
}

func (h *Handler) ListByLocation(c *gin.Context) {
	locationID := c.Param("id")

	// Parse filters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filters := store.ListShiftsFilters{
		LocationID: locationID,
	}

	if v := c.Query("shift_number"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filters.ShiftNumber = &n
		}
	}

	if v := c.Query("date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filters.ShiftDate = &t
		}
	}

	if v := c.Query("date_from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filters.DateFrom = &t
		}
	}

	if v := c.Query("date_to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filters.DateTo = &t
		}
	}

	shifts, total, err := h.store.ListShifts(c.Request.Context(), filters, limit, offset)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.OK(c, gin.H{
		"items": shifts,
		"meta":  response.Meta{Limit: limit, Offset: offset, Total: total},
	})
}

func (h *Handler) GetCurrent(c *gin.Context) {
	locationID := c.Param("id")

	shift, err := h.store.GetCurrentShift(c.Request.Context(), locationID)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "no shift configuration found for current time")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.OK(c, shift)
}

func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")
	shift, err := h.store.GetShiftByID(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "shift")
			return
		}
		response.InternalServerError(c)
		return
	}

	response.OK(c, shift)
}

func (h *Handler) GetSummary(c *gin.Context) {
	id := c.Param("id")

	summary, err := h.store.GetShiftSummary(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "shift")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.OK(c, summary)
}

func (h *Handler) logAudit(c *gin.Context, action, entityID string, locationID *string, metadata map[string]interface{}) {
	actorID := middleware.GetUserID(c)
	roleName := middleware.GetRoleName(c)
	ip := c.ClientIP()
	_ = h.store.CreateAuditLog(c.Request.Context(), store.AuditLogEntry{
		Action:     action,
		ActorID:    &actorID,
		ActorRole:  &roleName,
		EntityType: "shift",
		EntityID:   entityID,
		LocationID: locationID,
		IPAddress:  &ip,
		Metadata:   metadata,
	})
}
