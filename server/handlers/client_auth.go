package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"yesok-vietnam/server/models"
	"yesok-vietnam/server/pkg/jwt"
)

// ClientWechatLoginRequest 微信小程序登录请求。
type ClientWechatLoginRequest struct {
	Code      string `json:"code" binding:"required"`
	PhoneCode string `json:"phone_code"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

// WechatSessionResp 微信 jscode2session 接口返回结构。
type WechatSessionResp struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// exchangeWechatCode 调用微信接口将 code 换取 openid。
// 开发环境（ENV=dev）若无配置则返回模拟 openid。
func exchangeWechatCode(code string) (openid string, unionid string, err error) {
	appid := os.Getenv("WECHAT_APPID")
	secret := os.Getenv("WECHAT_SECRET")
	env := os.Getenv("ENV")

	// 开发环境兜底：允许模拟 openid
	if env == "dev" && (appid == "" || secret == "") {
		log.Printf("[Wechat] ENV=dev, returning mock openid for code: %s", code)
		return "dev_openid_" + code, "", nil
	}

	if appid == "" || secret == "" {
		return "", "", fmt.Errorf("WECHAT_APPID or WECHAT_SECRET not configured")
	}

	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appid, secret, code,
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", "", fmt.Errorf("failed to call wechat api: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read wechat response: %w", err)
	}

	var session WechatSessionResp
	if err := json.Unmarshal(body, &session); err != nil {
		return "", "", fmt.Errorf("failed to parse wechat response: %w", err)
	}

	if session.ErrCode != 0 {
		return "", "", fmt.Errorf("wechat error %d: %s", session.ErrCode, session.ErrMsg)
	}

	return session.OpenID, session.UnionID, nil
}

// ClientWechatLogin 处理 C 端微信小程序登录。
// 1.意图 -> 接收微信 wx.login() 返回的 code，换取 openid 并签发 JWT。
// 2.步骤 -> 调用微信接口查询/创建 app_users，更新用户资料，签发 JWT。
// 3.返回 -> token、user 信息和过期时间。
func ClientWechatLogin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ClientWechatLoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[WechatLogin] ShouldBindJSON failed: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "detail": err.Error()})
			return
		}

		log.Printf("[WechatLogin] code=%s nickname=%s", req.Code, req.Nickname)

		// 换取 openid
		openid, unionid, err := exchangeWechatCode(req.Code)
		if err != nil {
			log.Printf("[WechatLogin] exchangeWechatCode failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "微信登录失败，请稍后重试", "detail": err.Error()})
			return
		}
		log.Printf("[WechatLogin] openid=%s unionid=%s", openid, unionid)

		// 查询或创建 app_user
		var appUser models.AppUser
		result := db.Where("wechat_open_id = ?", openid).First(&appUser)

		if result.Error == gorm.ErrRecordNotFound {
			// 新用户：创建记录
			appUser = models.AppUser{
				WechatOpenID: openid,
				Nickname:     strings.TrimSpace(req.Nickname),
				AvatarURL:    strings.TrimSpace(req.AvatarURL),
				VipLevel:     1,
			}
			if appUser.Nickname == "" {
				appUser.Nickname = "微信用户"
			}
			if err := db.Create(&appUser).Error; err != nil {
				log.Printf("[WechatLogin] db.Create failed: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
				return
			}
			log.Printf("[WechatLogin] new user created | id=%d openid=%s", appUser.ID, openid)
		} else if result.Error != nil {
			log.Printf("[WechatLogin] db.First failed: %v", result.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		} else {
			// 老用户：同步昵称和头像
			updates := map[string]interface{}{}
			if req.Nickname != "" {
				updates["nickname"] = strings.TrimSpace(req.Nickname)
			}
			if req.AvatarURL != "" {
				updates["avatar_url"] = strings.TrimSpace(req.AvatarURL)
			}
			if len(updates) > 0 {
				db.Model(&appUser).Updates(updates)
				log.Printf("[WechatLogin] existing user | id=%d openid=%s — synced profile", appUser.ID, openid)
			} else {
				log.Printf("[WechatLogin] existing user | id=%d openid=%s — no profile update", appUser.ID, openid)
			}
		}

		// 签发 JWT（uid = app_users.id，role = RoleUser）
		jwtToken, expireUnix, err := jwt.Sign(appUser.ID, models.RoleUser)
		if err != nil {
			log.Printf("[WechatLogin] jwt.Sign failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
			return
		}

		log.Printf("[WechatLogin] success | user_id=%d token_prefix=%.10s... expire=%d",
			appUser.ID, jwtToken, expireUnix)

		c.JSON(http.StatusOK, gin.H{
			"token":  jwtToken,
			"expire": expireUnix,
			"user": gin.H{
				"id":             appUser.ID,
				"wechat_open_id": appUser.WechatOpenID,
				"nickname":       appUser.Nickname,
				"avatar_url":     appUser.AvatarURL,
				"phone":          appUser.Phone,
				"vip_level":      appUser.VipLevel,
			},
		})
	}
}
