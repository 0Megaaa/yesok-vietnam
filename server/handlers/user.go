package handlers

import (
	"net/http"
	"strings"

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
				"id":              appUser.ID,
				"wechat_open_id":  appUser.WechatOpenID,
				"nickname":        appUser.Nickname,
				"avatar_url":      appUser.AvatarURL,
				"phone":           appUser.Phone,
				"vip_level":       appUser.VipLevel,
				"balance":         appUser.Balance,
				"login_provider":  appUser.LoginProvider,
				"client_platform": appUser.ClientPlatform,
			},
		})
	}
}

// ClientUpdateProfile 更新 C 端用户资料（昵称、头像）。
func ClientUpdateProfile(db *gorm.DB) gin.HandlerFunc {
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

		var req struct {
			Nickname  string `json:"nickname"`
			AvatarURL string `json:"avatar_url"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		updates := map[string]interface{}{}
		if strings.TrimSpace(req.Nickname) != "" {
			updates["nickname"] = strings.TrimSpace(req.Nickname)
		}
		if strings.TrimSpace(req.AvatarURL) != "" {
			updates["avatar_url"] = strings.TrimSpace(req.AvatarURL)
		}

		if len(updates) > 0 {
			if err := db.Model(&models.AppUser{}).Where("id = ?", uid).Updates(updates).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
				return
			}
		}

		var appUser models.AppUser
		if err := db.First(&appUser, uid).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reload profile"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user": gin.H{
				"id":              appUser.ID,
				"wechat_open_id":  appUser.WechatOpenID,
				"nickname":        appUser.Nickname,
				"avatar_url":      appUser.AvatarURL,
				"phone":           appUser.Phone,
				"vip_level":       appUser.VipLevel,
				"balance":         appUser.Balance,
				"login_provider":  appUser.LoginProvider,
				"client_platform": appUser.ClientPlatform,
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
