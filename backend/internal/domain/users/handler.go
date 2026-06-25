package users

import (
	"strconv"

	"github.com/gin-gonic/gin"
	authsvc "github.com/thoriqzs/PARKIR/backend/internal/auth"
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

type CreateUserRequest struct {
	Name        string   `json:"name" binding:"required"`
	Email       string   `json:"email" binding:"required,email"`
	Password    string   `json:"password" binding:"required,min=8"`
	RoleID      string   `json:"role_id" binding:"required,uuid"`
	LocationIDs []string `json:"location_ids"`
}

type UpdateUserRequest struct {
	Name        *string   `json:"name,omitempty"`
	Email       *string   `json:"email,omitempty"`
	RoleID      *string   `json:"role_id,omitempty"`
	LocationIDs []string  `json:"location_ids,omitempty"`
	Status      *string   `json:"status,omitempty"`
}

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type ResetPINRequest struct {
	NewPIN string `json:"new_pin" binding:"required,len=6"`
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	users := r.Group("/users")
	users.Use(middleware.RequirePermission("users:view"))
	{
		users.GET("", h.List)
		users.GET("/:id", h.Get)
	}

	usersWithCreate := r.Group("/users")
	usersWithCreate.Use(middleware.RequirePermission("users:create"))
	{
		usersWithCreate.POST("", h.Create)
	}

	usersWithEdit := r.Group("/users")
	usersWithEdit.Use(middleware.RequirePermission("users:edit"))
	{
		usersWithEdit.PATCH("/:id", h.Update)
		usersWithEdit.POST("/:id/reset-password", h.ResetPassword)
		usersWithEdit.POST("/:id/reset-pin", h.ResetPIN)
	}

	usersWithDeactivate := r.Group("/users")
	usersWithDeactivate.Use(middleware.RequirePermission("users:deactivate"))
	{
		usersWithDeactivate.POST("/:id/deactivate", h.Deactivate)
	}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	passwordHash, err := authsvc.HashPassword(req.Password)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	user, err := h.store.CreateUser(c.Request.Context(), store.CreateUserInput{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: passwordHash,
		RoleID:       req.RoleID,
		LocationIDs:  req.LocationIDs,
	})
	if err != nil {
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "user.created", user.ID, nil, nil)
	response.Created(c, user)
}

func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, total, err := h.store.ListUsers(c.Request.Context(), limit, offset)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.OK(c, response.Response{
		Data: users,
		Meta: response.Meta{Limit: limit, Offset: offset, Total: total},
	})
}

func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")
	user, err := h.store.GetUserByID(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "user")
			return
		}
		response.InternalServerError(c)
		return
	}

	response.OK(c, user)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	user, err := h.store.UpdateUser(c.Request.Context(), id, store.UpdateUserInput{
		Name:        req.Name,
		Email:       req.Email,
		RoleID:      req.RoleID,
		LocationIDs: req.LocationIDs,
		Status:      req.Status,
	})
	if err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "user")
			return
		}
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "user.updated", user.ID, nil, gin.H{"fields": "name,email,role_id,location_ids,status"})
	response.OK(c, user)
}

func (h *Handler) ResetPassword(c *gin.Context) {
	id := c.Param("id")

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	passwordHash, err := authsvc.HashPassword(req.NewPassword)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	if err := h.store.UpdatePassword(c.Request.Context(), id, passwordHash); err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "user")
			return
		}
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "user.password_reset", id, nil, nil)
	response.OK(c, gin.H{"message": "password reset successfully"})
}

func (h *Handler) ResetPIN(c *gin.Context) {
	id := c.Param("id")

	var req ResetPINRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	pinHash, err := authsvc.HashPIN(req.NewPIN)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	if err := h.store.UpdatePIN(c.Request.Context(), id, pinHash); err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "user")
			return
		}
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "user.pin_reset", id, nil, nil)
	response.OK(c, gin.H{"message": "pin reset successfully"})
}

func (h *Handler) Deactivate(c *gin.Context) {
	id := c.Param("id")

	if err := h.store.DeactivateUser(c.Request.Context(), id); err != nil {
		if err == errors.ErrNotFound {
			response.NotFound(c, "user")
			return
		}
		response.InternalServerError(c)
		return
	}

	h.logAudit(c, "user.deactivated", id, nil, nil)
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
		EntityType: "user",
		EntityID:   entityID,
		LocationID: locationID,
		IPAddress:  &ip,
		Metadata:   metadata,
	})
}

// HasPermission helper for handler-level checks
func HasPermission(c *gin.Context, perm string) bool {
	return permissions.Has(middleware.GetPermissions(c), perm)
}
