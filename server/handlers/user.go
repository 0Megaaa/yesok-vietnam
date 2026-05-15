package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"yesok-vietnam/server/models"
)

// GetMe 返回已认证 C 端旧版用户资料。
// 1.意图 -> 保留原有客户端鉴权组件和 Telegram 登录链路的兼容能力。
// 2.步骤 -> 从 JWT 中间件读取 uid，再查询旧版 users 表并组装安全字段。
// 3.返回 -> 客户端个人资料 JSON；若未登录或用户不存在则返回错误。
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
