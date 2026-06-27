package alerts

import (
	"strconv"

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

type AcknowledgeRequest struct {
	ResolutionNotes *string `json:"resolution_notes,omitempty"`
}

type ResolveRequest struct {
	ResolutionNotes string `json:"resolution_notes" binding:"required"`
}

type UpdateAlertConfigRequest struct {
	Enabled   *bool                   `json:"enabled,omitempty"`
	Threshold *map[string]interface{} `json:"threshold,omitempty"`
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	alerts := r.Group("/alerts")
	alerts.Use(middleware.RequirePermission("observability:view_alerts"))
	{
		alerts.GET("", h.List)
		alerts.GET("/:id", h.Get)
		alerts.POST("/:id/acknowledge", h.Acknowledge)
		alerts.POST("/:id/resolve", h.Resolve)
	}

	configs := r.Group("/alert-configs")
	configs.Use(middleware.RequirePermission("observability:view_alerts"))
	{
		configs.GET("", h.ListConfigs)
	}

	configsManage := r.Group("/alert-configs")
	configsManage.Use(middleware.RequirePermission("observability:manage_alerts"))
	{
		configsManage.PATCH("/:id", h.UpdateConfig)
	}
}

func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filters := store.ListAlertsFilters{
		LocationID: c.Query("location_id"),
		State:      c.Query("state"),
		Code:       c.Query("code"),
	}

	alerts, total, err := h.store.ListAlerts(c.Request.Context(), filters, limit, offset)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	triggeredCount, _ := h.store.CountTriggeredAlerts(c.Request.Context())

	response.OK(c, gin.H{
		"items": alerts,
		"meta": response.Response{
			Data: map[string]interface{}{
				"limit":           limit,
				"offset":          offset,
				"total":           total,
				"triggered_count": triggeredCount,
			},
		},
	})
}

func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")
	a, err := h.store.GetAlertByID(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "alert")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}
	response.OK(c, a)
}

func (h *Handler) Acknowledge(c *gin.Context) {
	id := c.Param("id")
	actorID := middleware.GetUserID(c)

	a, err := h.store.AcknowledgeAlert(c.Request.Context(), id, actorID)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "alert")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}
	response.OK(c, a)
}

func (h *Handler) Resolve(c *gin.Context) {
	id := c.Param("id")

	var req ResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	actorID := middleware.GetUserID(c)

	a, err := h.store.ResolveAlert(c.Request.Context(), id, actorID, req.ResolutionNotes)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "alert")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}
	response.OK(c, a)
}

func (h *Handler) ListConfigs(c *gin.Context) {
	locationID := c.Query("location_id")
	configs, err := h.store.ListAlertConfigs(c.Request.Context(), locationID)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}
	response.OK(c, configs)
}

func (h *Handler) UpdateConfig(c *gin.Context) {
	id := c.Param("id")

	var req UpdateAlertConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	actorID := middleware.GetUserID(c)

	config, err := h.store.UpdateAlertConfig(c.Request.Context(), id, actorID, req.Enabled, *req.Threshold)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "alert config")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}
	response.OK(c, config)
}