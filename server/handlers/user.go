package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"yesok-vietnam/server/models"
)

// ClientMe 返回已认证 C 端用户资料（基于 app_users 表）。
// 1.意图 -> 为微信小程序等 C 端用户提供个人资料查询接口。
// 2.步骤 -> 从 JWT 中间件读取 uid，再查询 app_users 表并组装安全字段。
// 3.返回 -> 客户端个人资料 JSON；若未登录或用户不存在则返回错误。
func ClientMe(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidVal, ok := c.Get("uid")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		uid, ok := uidVal.(uint)
		if !ok || uid == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var appUser models.AppUser
		if err := db.First(&appUser, uid).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user": gin.H{
				"id":         appUser.ID,
				"nickname":   appUser.Nickname,
				"avatar_url": appUser.AvatarURL,
				"phone":      appUser.Phone,
				"vip_level":  appUser.VipLevel,
				"balance":    appUser.Balance,
			},
		})
	}
}

// GetMe 返回已认证 C 端旧版用户资料（基于 users 表，保留兼容性）。
// Deprecated: 新代码请使用 ClientMe。
func GetMe(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, _ := c.Get("uid")
		uidVal, ok := uid.(uint)
		if !ok || uidVal == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var user models.User
		if err := db.First(&user, uidVal).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
			"balance":    user.Balance,
			"language":   user.Language,
			"phone":      "",
			"avatar_url": user.AvatarURL,
		})
	}
}
