package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"yesok-vietnam/pkg/jwt"
)

// AuthMiddleware is stateless — it only reads the JWT and injects claims into
// the Gin context.  Handlers that need the full User model load it from the
// DB using the UID from context.
type AuthMiddleware struct{}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

// RequireAuth validates the JWT Bearer token and injects claims into the context.
// Used by both /api/v1/client and /api/v1/admin route groups.
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization format, expected: Bearer <token>"})
			return
		}

		claims, err := jwt.Validate(token)
		if err != nil {
			if err == jwt.ErrTokenExpired {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token has expired"})
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Inject claims into context.
		c.Set("uid", claims.UID)
		c.Set("role", claims.Role)
		c.Set("isAdmin", claims.IsAdmin)

		c.Next()
	}
}

// RequireRole returns a middleware that restricts access to users with one of the given roles.
// Must be applied AFTER RequireAuth in the middleware chain.
func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		role := roleVal.(string)
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}

// GetUID reads the authenticated user ID from the context.
func GetUID(c *gin.Context) uint {
	if v, ok := c.Get("uid"); ok {
		return v.(uint)
	}
	return 0
}

// GetRole reads the authenticated user's role from the context.
func GetRole(c *gin.Context) string {
	if v, ok := c.Get("role"); ok {
		return v.(string)
	}
	return ""
}

// IsAdmin returns true if the current request is from an admin.
func IsAdmin(c *gin.Context) bool {
	if v, ok := c.Get("isAdmin"); ok {
		return v.(bool)
	}
	return false
}
