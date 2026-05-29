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
	Code           string `json:"code" binding:"required"`
	PhoneCode      string `json:"phone_code"`
	Nickname       string `json:"nickname"`
	AvatarURL      string `json:"avatar_url"`
	LoginProvider  string `json:"login_provider"`
	ClientPlatform string `json:"client_platform"`
	DevIdentity    string `json:"dev_identity"`
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
// 注意：wx.login 的 code 是一次性临时凭证，每次都会变化，不能作为用户唯一标识。
// 本地未配置 WECHAT_APPID/WECHAT_SECRET 时，使用前端持久化 dev_identity 生成稳定 mock openid。
// 生产环境必须使用微信 jscode2session 返回的真实 openid。
func exchangeWechatCode(code string, devIdentity string) (openid string, unionid string, err error) {
	appid := os.Getenv("WECHAT_APPID")
	secret := os.Getenv("WECHAT_SECRET")
	env := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))

	if appid == "" || secret == "" {
		if env == "dev" || env == "local" || env == "test" || env == "" {
			stableID := strings.TrimSpace(devIdentity)
			if stableID == "" {
				stableID = code
				log.Printf("[Wechat] WARNING: dev_identity missing, fallback to login code; mock openid may change every login")
			}

			replacer := strings.NewReplacer(
				" ", "_",
				"/", "_",
				"\\", "_",
				":", "_",
				"\n", "_",
				"\r", "_",
				"\t", "_",
			)
			stableID = replacer.Replace(stableID)

			mockOpenID := "dev_openid_" + stableID
			log.Printf("[Wechat] ENV=%s, WECHAT_APPID/SECRET missing, using stable mock openid: %s", env, mockOpenID)
			return mockOpenID, "", nil
		}

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

		log.Printf("[WechatLogin] code=%s nickname=%s dev_identity=%s",
			req.Code, req.Nickname, req.DevIdentity)

		// 换取 openid
		openid, unionid, err := exchangeWechatCode(req.Code, req.DevIdentity)
		if err != nil {
			log.Printf("[WechatLogin] exchangeWechatCode failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "微信登录失败，请稍后重试", "detail": err.Error()})
			return
		}
		log.Printf("[WechatLogin] openid=%s unionid=%s", openid, unionid)

		// 平台来源默认值
		loginProvider := strings.TrimSpace(req.LoginProvider)
		if loginProvider == "" {
			loginProvider = "wechat"
		}
		clientPlatform := strings.TrimSpace(req.ClientPlatform)
		if clientPlatform == "" {
			clientPlatform = "mp_weixin"
		}
		log.Printf("[WechatLogin] provider=%s platform=%s", loginProvider, clientPlatform)

		// 查询或创建 app_user（老数据仍按 wechat_open_id 查询，不走 login_provider）
		var appUser models.AppUser
		result := db.Where("wechat_open_id = ?", openid).First(&appUser)

		if result.Error == gorm.ErrRecordNotFound {
			// 新用户：创建记录
			appUser = models.AppUser{
				WechatOpenID:   openid,
				Nickname:       strings.TrimSpace(req.Nickname),
				AvatarURL:      strings.TrimSpace(req.AvatarURL),
				LoginProvider:  loginProvider,
				ClientPlatform: clientPlatform,
				VipLevel:       1,
			}
			if appUser.Nickname == "" {
				appUser.Nickname = "微信用户"
			}
			if err := db.Create(&appUser).Error; err != nil {
				log.Printf("[WechatLogin] db.Create failed: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
				return
			}
			log.Printf("[WechatLogin] new user created | id=%d openid=%s login_provider=%s client_platform=%s",
				appUser.ID, openid, loginProvider, clientPlatform)
		} else if result.Error != nil {
			log.Printf("[WechatLogin] db.First failed: %v", result.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		} else {
			// 老用户：同步昵称、头像和平台字段
			updates := map[string]interface{}{
				"login_provider":  loginProvider,
				"client_platform": clientPlatform,
			}
			if strings.TrimSpace(req.Nickname) != "" {
				updates["nickname"] = strings.TrimSpace(req.Nickname)
			}
			if strings.TrimSpace(req.AvatarURL) != "" {
				updates["avatar_url"] = strings.TrimSpace(req.AvatarURL)
			}
			if err := db.Model(&appUser).Updates(updates).Error; err != nil {
				log.Printf("[WechatLogin] db.Updates failed: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
				return
			}
			// 重新查询保证返回最新字段
			db.First(&appUser, appUser.ID)
			log.Printf("[WechatLogin] existing user | id=%d openid=%s — synced profile & platform", appUser.ID, openid)
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
