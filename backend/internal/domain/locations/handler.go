package locations

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

type CreateLocationRequest struct {
	Name     string                 `json:"name" binding:"required"`
	Code     string                 `json:"code" binding:"required"`
	Address  string                 `json:"address"`
	City     string                 `json:"city"`
	Capacity map[string]interface{} `json:"capacity"`
}

type UpdateLocationRequest struct {
	Name     *string                `json:"name,omitempty"`
	Address  *string                `json:"address,omitempty"`
	City     *string                `json:"city,omitempty"`
	Status   *string                `json:"status,omitempty"`
	Capacity map[string]interface{} `json:"capacity,omitempty"`
}

type AssignOperatorRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	locations := r.Group("/locations")
	locations.Use(middleware.RequirePermission("locations:view"))
	{
		locations.GET("", h.List)
		locations.GET("/:id", h.Get)
	}

	locationsWithManage := r.Group("/locations")
	locationsWithManage.Use(middleware.RequirePermission("locations:create"))
	{
		locationsWithManage.POST("", h.Create)
		locationsWithManage.PATCH("/:id", h.Update)
		locationsWithManage.POST("/:id/deactivate", h.Deactivate)
	}

	locationsWithAssign := r.Group("/locations")
	locationsWithAssign.Use(middleware.RequirePermission("locations:assign_operators"))
	{
		locationsWithAssign.POST("/:id/assign-operator", h.AssignOperator)
		locationsWithAssign.POST("/:id/remove-operator", h.RemoveOperator)
	}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	loc, err := h.store.CreateLocation(c.Request.Context(), store.CreateLocationInput{
		Name:     req.Name,
		Code:     req.Code,
		Address:  req.Address,
		City:     req.City,
		Capacity: req.Capacity,
	})
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "location.created", loc.ID, &loc.ID, gin.H{"name": req.Name, "code": req.Code})
	response.Created(c, loc)
}

func (h *Handler) List(c *gin.Context) {
	locations, total, err := h.store.ListLocations(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	response.OK(c, gin.H{
		"items": locations,
		"meta":  response.Meta{Limit: total, Offset: 0, Total: total},
	})
}

func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")
	loc, err := h.store.GetLocationByID(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "location")
			return
		}
		response.InternalServerError(c)
		return
	}

	response.OK(c, loc)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	loc, err := h.store.UpdateLocation(c.Request.Context(), id, store.UpdateLocationInput{
		Name:     req.Name,
		Address:  req.Address,
		City:     req.City,
		Status:   req.Status,
		Capacity: req.Capacity,
	})
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "location")
			return
		}
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "location.updated", loc.ID, &loc.ID, gin.H{"name": req.Name, "status": req.Status})
	response.OK(c, loc)
}

func (h *Handler) Deactivate(c *gin.Context) {
	id := c.Param("id")

	loc, err := h.store.DeactivateLocation(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "location")
			return
		}
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "location.deactivated", loc.ID, &loc.ID, nil)
	response.OK(c, loc)
}

func (h *Handler) AssignOperator(c *gin.Context) {
	locationID := c.Param("id")

	var req AssignOperatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	if err := h.store.AssignOperatorToLocation(c.Request.Context(), locationID, req.UserID); err != nil {
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "location.operator_assigned", locationID, &locationID, gin.H{"user_id": req.UserID})
	response.NoContent(c)
}

func (h *Handler) RemoveOperator(c *gin.Context) {
	locationID := c.Param("id")

	var req AssignOperatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	if err := h.store.RemoveOperatorFromLocation(c.Request.Context(), locationID, req.UserID); err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "assignment")
			return
		}
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "location.operator_removed", locationID, &locationID, gin.H{"user_id": req.UserID})
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
		EntityType: "location",
		EntityID:   entityID,
		LocationID: locationID,
		IPAddress:  &ip,
		Metadata:   metadata,
	})
}
