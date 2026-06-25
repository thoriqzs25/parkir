package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/thoriqzs/PARKIR/backend/internal/auth"
	"github.com/thoriqzs/PARKIR/backend/internal/permissions"
	"github.com/thoriqzs/PARKIR/backend/internal/response"
)

type contextKey string

const (
	ContextUserIDKey   contextKey = "user_id"
	ContextEmailKey    contextKey = "email"
	ContextRoleIDKey   contextKey = "role_id"
	ContextRoleNameKey contextKey = "role_name"
	ContextPermsKey    contextKey = "permissions"
)

func Auth(authService *auth.Service, permResolver *permissions.Resolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ""

		cookie, err := c.Cookie("access_token")
		if err == nil && cookie != "" {
			tokenString = cookie
			if strings.HasPrefix(tokenString, "Bearer ") {
				tokenString = strings.TrimPrefix(tokenString, "Bearer ")
			}
		} else {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenString == "" {
			response.Unauthorized(c, "missing access token")
			c.Abort()
			return
		}

		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(string(ContextUserIDKey), claims.UserID)
		c.Set(string(ContextEmailKey), claims.Email)
		c.Set(string(ContextRoleIDKey), claims.RoleID)
		c.Set(string(ContextRoleNameKey), claims.RoleName)

		var locationID *string
		if locID := c.Param("location_id"); locID != "" {
			locationID = &locID
		} else if locID := c.Query("location_id"); locID != "" {
			locationID = &locID
		}

		perms, err := permResolver.EffectivePermissions(c.Request.Context(), claims.UserID, locationID)
		if err != nil {
			response.InternalServerError(c)
			c.Abort()
			return
		}

		c.Set(string(ContextPermsKey), perms)

		c.Next()
	}
}

func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		perms := GetPermissions(c)
		if !permissions.Has(perms, permission) {
			response.Forbidden(c, "missing required permission: "+permission)
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireAnyPermission(permsList ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		perms := GetPermissions(c)
		if !permissions.HasAny(perms, permsList...) {
			response.Forbidden(c, "missing required permission")
			c.Abort()
			return
		}
		c.Next()
	}
}

func GetUserID(c *gin.Context) string {
	val, _ := c.Get(string(ContextUserIDKey))
	if id, ok := val.(string); ok {
		return id
	}
	return ""
}

func GetEmail(c *gin.Context) string {
	val, _ := c.Get(string(ContextEmailKey))
	if email, ok := val.(string); ok {
		return email
	}
	return ""
}

func GetRoleName(c *gin.Context) string {
	val, _ := c.Get(string(ContextRoleNameKey))
	if name, ok := val.(string); ok {
		return name
	}
	return ""
}

func GetPermissions(c *gin.Context) []string {
	val, _ := c.Get(string(ContextPermsKey))
	if perms, ok := val.([]string); ok {
		return perms
	}
	return nil
}

// Context helpers for use outside of gin handlers
func UserIDFromContext(ctx context.Context) string {
	if val := ctx.Value(ContextUserIDKey); val != nil {
		if id, ok := val.(string); ok {
			return id
		}
	}
	return ""
}
