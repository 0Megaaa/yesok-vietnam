package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"yesok-vietnam/server/models"
	"yesok-vietnam/server/pkg/jwt"
)

// AuthMiddleware validates JWT tokens and, for B-end roles (admin/manager),
// performs single sign-on verification against sys_users.current_token_hash.
type AuthMiddleware struct {
	db *gorm.DB
}

func NewAuthMiddleware(db *gorm.DB) *AuthMiddleware {
	return &AuthMiddleware{db: db}
}

// RequireAuth validates the JWT Bearer token and injects claims into the context.
// For B-end roles (admin, manager), it additionally verifies the token hash against
// sys_users.current_token_hash to enforce single sign-on.
// Used by both /api/v1/client and /api/v1/admin route groups.
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "TOKEN_MISSING", "error": "登录状态已失效，请重新登录"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "TOKEN_INVALID", "error": "登录状态已失效，请重新登录"})
			return
		}

		claims, err := jwt.Validate(token)
		if err != nil {
			if err == jwt.ErrTokenExpired {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "TOKEN_EXPIRED", "error": "登录状态已失效，请重新登录"})
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "TOKEN_INVALID", "error": "登录状态已失效，请重新登录"})
			return
		}

		// B-end single sign-on verification: admin or manager roles
		if claims.Role == models.RoleAdmin || claims.Role == models.RoleManager {
			var user models.SysUser
			if err := m.db.Where("id = ?", claims.UID).First(&user).Error; err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "ACCOUNT_DISABLED", "error": "登录状态已失效，请重新登录"})
				return
			}
			if user.Status != 1 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "ACCOUNT_DISABLED", "error": "登录状态已失效，请重新登录"})
				return
			}
			// Check token expiry from DB (authoritative)
			if user.TokenExpiresAt == nil || user.TokenExpiresAt.Before(time.Now()) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "TOKEN_EXPIRED", "error": "登录状态已失效，请重新登录"})
				return
			}
			// Verify token hash matches current session
			currentHash := jwt.TokenHash(token)
			if user.CurrentTokenHash == "" || user.CurrentTokenHash != currentHash {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "TOKEN_KICKED", "error": "登录状态已失效，请重新登录"})
				return
			}
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
