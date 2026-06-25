package roles

import (
	"github.com/gin-gonic/gin"
	"github.com/thoriqzs/PARKIR/backend/internal/errors"
	"github.com/thoriqzs/PARKIR/backend/internal/middleware"
	"github.com/thoriqzs/PARKIR/backend/internal/permissions"
	"github.com/thoriqzs/PARKIR/backend/internal/response"
	"github.com/thoriqzs/PARKIR/backend/internal/store"
)

type Handler struct {
	store *store.Store
}

func NewHandler(store *store.Store) *Handler {
	return &Handler{store: store}
}

type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required"`
	Permissions []string `json:"permissions" binding:"required"`
}

type UpdateRoleRequest struct {
	Name        *string   `json:"name,omitempty"`
	Permissions []string  `json:"permissions,omitempty"`
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	roles := r.Group("/roles")
	roles.Use(middleware.RequirePermission("users:view"))
	{
		roles.GET("", h.List)
		roles.GET("/:id", h.Get)
	}

	rolesWithManage := r.Group("/roles")
	rolesWithManage.Use(middleware.RequirePermission("users:create"))
	{
		rolesWithManage.POST("", h.Create)
		rolesWithManage.PATCH("/:id", h.Update)
		rolesWithManage.DELETE("/:id", h.Delete)
	}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	if err := permissions.ValidateList(req.Permissions); err != nil {
		response.BadRequest(c, "INVALID_PERMISSION", err.Error())
		return
	}

	if hasFinancePermission(req.Permissions) && middleware.GetRoleName(c) != "owner" {
		response.Forbidden(c, "only owners can grant finance permissions")
		return
	}

	role, err := h.store.CreateRole(c.Request.Context(), store.CreateRoleInput{
		Name:        req.Name,
		Permissions: permissions.Expand(req.Permissions),
	})
	if err != nil {
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "role.created", role.ID, gin.H{"name": req.Name, "permissions": req.Permissions})
	response.Created(c, role)
}

func (h *Handler) List(c *gin.Context) {
	roles, err := h.store.ListRoles(c.Request.Context())
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.OK(c, roles)
}

func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")
	role, err := h.store.GetRoleByID(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "role")
			return
		}
		response.InternalServerError(c)
		return
	}

	response.OK(c, role)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	if req.Permissions != nil {
		if err := permissions.ValidateList(req.Permissions); err != nil {
			response.BadRequest(c, "INVALID_PERMISSION", err.Error())
			return
		}

		if hasFinancePermission(req.Permissions) && middleware.GetRoleName(c) != "owner" {
			response.Forbidden(c, "only owners can grant finance permissions")
			return
		}

		expanded := permissions.Expand(req.Permissions)
		req.Permissions = expanded
	}

	role, err := h.store.UpdateRole(c.Request.Context(), id, store.UpdateRoleInput{
		Name:        req.Name,
		Permissions: req.Permissions,
	})
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "role")
			return
		}
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "role.updated", role.ID, gin.H{"name": req.Name, "permissions": req.Permissions})
	response.OK(c, role)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.store.SoftDeleteRole(c.Request.Context(), id); err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "role")
			return
		}
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "role.deleted", id, nil)
	response.NoContent(c)
}

func (h *Handler) logAudit(c *gin.Context, action, entityID string, metadata map[string]interface{}) {
	actorID := middleware.GetUserID(c)
	roleName := middleware.GetRoleName(c)
	ip := c.ClientIP()
	_ = h.store.CreateAuditLog(c.Request.Context(), store.AuditLogEntry{
		Action:     action,
		ActorID:    &actorID,
		ActorRole:  &roleName,
		EntityType: "role",
		EntityID:   entityID,
		IPAddress:  &ip,
		Metadata:   metadata,
	})
}

func hasFinancePermission(perms []string) bool {
	for _, p := range perms {
		if p == "finance:*" || len(p) > 8 && p[:8] == "finance:" {
			return true
		}
	}
	return false
}
