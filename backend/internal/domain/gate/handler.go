package gate

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

type RegisterGateRequest struct {
	DeviceID   string  `json:"device_id" binding:"required"`
	Name       string  `json:"name"`
	LocationID *string `json:"location_id,omitempty"`
	IPAddress  string  `json:"ip_address"`
}

type UpdateGateRequest struct {
	Name       *string `json:"name,omitempty"`
	LocationID *string `json:"location_id,omitempty"`
	IPAddress  *string `json:"ip_address,omitempty"`
}

func (h *Handler) RegisterPublicRoutes(r *gin.RouterGroup) {
	r.GET("/gate/:id/info", h.GetGateInfo)
}

func (h *Handler) RegisterAdminRoutes(r *gin.RouterGroup) {
	gates := r.Group("/gates")
	gates.Use(middleware.RequirePermission("gates:view"))
	{
		gates.GET("", h.ListGates)
		gates.GET("/:id", h.GetGate)
	}

	gatesCreate := r.Group("/gates")
	gatesCreate.Use(middleware.RequirePermission("gates:register"))
	{
		gatesCreate.POST("", h.RegisterGate)
	}

	gatesEdit := r.Group("/gates")
	gatesEdit.Use(middleware.RequirePermission("gates:edit"))
	{
		gatesEdit.PATCH("/:id", h.UpdateGate)
	}

	gatesDelete := r.Group("/gates")
	gatesDelete.Use(middleware.RequirePermission("gates:delete"))
	{
		gatesDelete.DELETE("/:id", h.DeleteGate)
	}
}

func (h *Handler) GetGateInfo(c *gin.Context) {
	id := c.Param("id")

	info, err := h.store.GetGateInfo(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "location")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.OK(c, info)
}

func (h *Handler) ListGates(c *gin.Context) {
	locationID := c.Query("location_id")

	gates, err := h.store.ListGates(c.Request.Context(), locationID)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.OK(c, gates)
}

func (h *Handler) GetGate(c *gin.Context) {
	id := c.Param("id")

	gate, err := h.store.GetGateByID(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrNotFound {
			// Fallback: try by device_id
			gate, err = h.store.GetGateByDeviceID(c.Request.Context(), id)
			if err != nil {
				if err == errors.ErrNotFound {
					response.NotFound(c, "gate")
					return
				}
				_ = c.Error(err)
				response.InternalServerError(c)
				return
			}
			response.OK(c, gate)
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.OK(c, gate)
}

func (h *Handler) RegisterGate(c *gin.Context) {
	var req RegisterGateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	gate, err := h.store.RegisterGate(c.Request.Context(), store.RegisterGateInput{
		DeviceID:   req.DeviceID,
		Name:       req.Name,
		LocationID: req.LocationID,
		IPAddress:  req.IPAddress,
	})
	if err != nil {
		if err == errors.ErrConflict {
			response.Conflict(c, "GATE_EXISTS", "device_id already registered")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.Created(c, gate)
}

func (h *Handler) UpdateGate(c *gin.Context) {
	id := c.Param("id")

	var req UpdateGateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	gate, err := h.store.UpdateGate(c.Request.Context(), id, store.UpdateGateInput{
		Name:       req.Name,
		LocationID: req.LocationID,
		IPAddress:  req.IPAddress,
	})
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "gate")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.OK(c, gate)
}

func (h *Handler) DeleteGate(c *gin.Context) {
	id := c.Param("id")

	err := h.store.DeleteGate(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "gate")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.OK(c, gin.H{"message": "gate deleted"})
}
