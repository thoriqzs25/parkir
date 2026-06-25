package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	authsvc "github.com/thoriqzs/PARKIR/backend/internal/auth"
	"github.com/thoriqzs/PARKIR/backend/internal/errors"
	"github.com/thoriqzs/PARKIR/backend/internal/response"
	"github.com/thoriqzs/PARKIR/backend/internal/store"
)

type Handler struct {
	authService *authsvc.Service
	store       *store.Store
}

func NewHandler(authService *authsvc.Service, store *store.Store) *Handler {
	return &Handler{
		authService: authService,
		store:       store,
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	User  *store.User `json:"user"`
	Token string      `json:"token"`
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", h.Login)
		authGroup.POST("/logout", h.Logout)
		authGroup.POST("/refresh", h.Refresh)
	}
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	user, err := h.store.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		if err == errors.ErrNotFound {
			response.Unauthorized(c, "invalid credentials")
			return
		}
		response.InternalServerError(c)
		return
	}

	if !authsvc.CheckPassword(req.Password, user.PasswordHash) {
		response.Unauthorized(c, "invalid credentials")
		return
	}

	token, err := h.authService.GenerateAccessToken(user.ID, user.Email, user.RoleID, user.RoleName)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"access_token",
		token,
		int(8*time.Hour.Seconds()),
		"/",
		"",
		false,
		true,
	)

	response.OK(c, LoginResponse{
		User:  user,
		Token: token,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	c.SetCookie(
		"access_token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)
	response.OK(c, gin.H{"message": "logged out"})
}

func (h *Handler) Refresh(c *gin.Context) {
	cookie, err := c.Cookie("access_token")
	if err != nil {
		response.Unauthorized(c, "missing access token")
		return
	}

	claims, err := h.authService.ValidateToken(cookie)
	if err != nil {
		response.Unauthorized(c, "invalid or expired token")
		return
	}

	newToken, err := h.authService.GenerateAccessToken(claims.UserID, claims.Email, claims.RoleID, claims.RoleName)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	c.SetCookie(
		"access_token",
		newToken,
		int(8*time.Hour.Seconds()),
		"/",
		"",
		false,
		true,
	)

	response.OK(c, gin.H{"token": newToken})
}
