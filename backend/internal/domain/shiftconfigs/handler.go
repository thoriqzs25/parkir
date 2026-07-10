package shiftconfigs

import (
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

type CreateShiftConfigRequest struct {
	ShiftCode   string `json:"shift_code" binding:"required"`
	ShiftNumber int    `json:"shift_number" binding:"required,min=1"`
	StartTime   string `json:"start_time" binding:"required,datetime=15:04:05"`
	EndTime     string `json:"end_time" binding:"required,datetime=15:04:05"`
}

type UpdateShiftConfigRequest struct {
	ShiftCode   *string `json:"shift_code,omitempty"`
	ShiftNumber *int    `json:"shift_number,omitempty"`
	StartTime   *string `json:"start_time,omitempty"`
	EndTime     *string `json:"end_time,omitempty"`
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	configs := r.Group("/locations/:id/shift-configs")
	configs.Use(middleware.RequirePermission("shifts:manage"))
	{
		configs.GET("", h.List)
		configs.POST("", h.Create)
		configs.PUT("/:code", h.Update)
		configs.DELETE("/:code", h.Delete)
	}
}

func (h *Handler) List(c *gin.Context) {
	locationID := c.Param("id")

	configs, err := h.store.ListLocationShiftConfigs(c.Request.Context(), locationID)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.OK(c, gin.H{
		"items": configs,
		"meta":  response.Meta{Limit: len(configs), Offset: 0, Total: len(configs)},
	})
}

func (h *Handler) Create(c *gin.Context) {
	locationID := c.Param("id")

	var req CreateShiftConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	config, err := h.store.CreateLocationShiftConfig(c.Request.Context(), store.CreateLocationShiftConfigInput{
		LocationID:  locationID,
		ShiftCode:   req.ShiftCode,
		ShiftNumber: req.ShiftNumber,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
	})
	if err != nil {
		// Check for overlap error
		if isOverlapError(err) {
			response.BadRequest(c, "OVERLAP_ERROR", "Shift configuration overlaps with existing shift")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "shift_config.created", config.ID, &locationID, gin.H{
		"shift_code":   req.ShiftCode,
		"shift_number": req.ShiftNumber,
		"start_time":   req.StartTime,
		"end_time":     req.EndTime,
	})

	response.Created(c, config)
}

func (h *Handler) Update(c *gin.Context) {
	locationID := c.Param("id")
	code := c.Param("code")

	// Get existing config by code
	existing, err := h.store.GetLocationShiftConfigByCode(c.Request.Context(), locationID, code)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "shift config")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	var req UpdateShiftConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	config, err := h.store.UpdateLocationShiftConfig(c.Request.Context(), existing.ID, store.UpdateLocationShiftConfigInput{
		ShiftCode:   req.ShiftCode,
		ShiftNumber: req.ShiftNumber,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
	})
	if err != nil {
		if isOverlapError(err) {
			response.BadRequest(c, "OVERLAP_ERROR", "Shift configuration overlaps with existing shift")
			return
		}
		if err == errors.ErrNotFound {
			response.NotFound(c, "shift config")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "shift_config.updated", config.ID, &locationID, gin.H{
		"old_code":     code,
		"new_code":     req.ShiftCode,
		"shift_number": req.ShiftNumber,
	})

	response.OK(c, config)
}

func (h *Handler) Delete(c *gin.Context) {
	locationID := c.Param("id")
	code := c.Param("code")

	// Get existing config by code
	existing, err := h.store.GetLocationShiftConfigByCode(c.Request.Context(), locationID, code)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "shift config")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	if err := h.store.DeleteLocationShiftConfig(c.Request.Context(), existing.ID); err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "shift config")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "shift_config.deleted", existing.ID, &locationID, gin.H{
		"shift_code": code,
	})

	response.NoContent(c)
}

func (h *Handler) logAudit(c *gin.Context, action, entityID string, locationID *string, metadata map[string]interface{}) {
	actorID := middleware.GetUserID(c)
	roleName := middleware.GetRoleName(c)
	ip := c.ClientIP()
	_ = h.store.CreateAuditLog(c.Request.Context(), store.AuditLogEntry{
		Action:     action,
		ActorID:    &actorID,
		ActorRole:  &roleName,
		EntityType: "shift_config",
		EntityID:   entityID,
		LocationID: locationID,
		IPAddress:  &ip,
		Metadata:   metadata,
	})
}

func isOverlapError(err error) bool {
	if err == nil {
		return false
	}
	// Check for PostgreSQL overlap error from trigger
	errStr := err.Error()
	return containsStr(errStr, "Overlapping shift configuration")
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStrHelper(s, substr))
}

func containsStrHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
