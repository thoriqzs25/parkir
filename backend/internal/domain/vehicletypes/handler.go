package vehicletypes

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

type CreateVehicleTypeRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=20"`
	DisplayName string `json:"display_name" binding:"required,min=1,max=100"`
	Description string `json:"description"`
}

type UpdateVehicleTypeRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	vtsWithView := r.Group("/vehicle-types")
	vtsWithView.Use(middleware.RequirePermission("vehicle-types:view"))
	{
		vtsWithView.GET("", h.List)
		vtsWithView.GET("/:name", h.Get)
	}

	vtsWithCreate := r.Group("/vehicle-types")
	vtsWithCreate.Use(middleware.RequirePermission("vehicle-types:create"))
	{
		vtsWithCreate.POST("", h.Create)
	}

	vtsWithEdit := r.Group("/vehicle-types")
	vtsWithEdit.Use(middleware.RequirePermission("vehicle-types:edit"))
	{
		vtsWithEdit.PATCH("/:name", h.Update)
	}

	vtsWithDelete := r.Group("/vehicle-types")
	vtsWithDelete.Use(middleware.RequirePermission("vehicle-types:delete"))
	{
		vtsWithDelete.DELETE("/:name", h.Delete)
	}
}

func (h *Handler) List(c *gin.Context) {
	types, err := h.store.ListVehicleTypes(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}
	response.OK(c, types)
}

func (h *Handler) Get(c *gin.Context) {
	name := c.Param("name")

	vt, err := h.store.GetVehicleType(c.Request.Context(), name)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "vehicle type")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}
	response.OK(c, vt)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateVehicleTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	vt, err := h.store.CreateVehicleType(c.Request.Context(), store.CreateVehicleTypeInput{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
	})
	if err != nil {
		if err == errors.ErrConflict {
			response.Conflict(c, "VEHICLE_TYPE_EXISTS", err.Error())
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.Created(c, vt)
}

func (h *Handler) Update(c *gin.Context) {
	name := c.Param("name")

	var req UpdateVehicleTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	vt, err := h.store.UpdateVehicleType(c.Request.Context(), name, store.UpdateVehicleTypeInput{
		DisplayName: req.DisplayName,
		Description: req.Description,
	})
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "vehicle type")
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.OK(c, vt)
}

func (h *Handler) Delete(c *gin.Context) {
	name := c.Param("name")

	err := h.store.DeleteVehicleType(c.Request.Context(), name)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "vehicle type")
			return
		}
		if err == errors.ErrInUse {
			response.Conflict(c, "VEHICLE_TYPE_IN_USE", err.Error())
			return
		}
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.OK(c, gin.H{"message": "vehicle type deleted"})
}
