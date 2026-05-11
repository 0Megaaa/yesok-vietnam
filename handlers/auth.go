package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"yesok-vietnam/models"
	"yesok-vietnam/pkg/telegram"
)

const (
	MaxAuthAge = 24 * time.Hour
	TokenTTL   = 7 * 24 * time.Hour
)

type AuthTGRequest struct {
	InitData string `json:"initData" binding:"required"`
}

// AuthTGResponse is the JSON returned to the frontend.
// Only client-facing fields are included (no DB internals like session_token, deleted_at, etc.).
type AuthTGResponse struct {
	Token  string           `json:"token"`
	User   *AuthUserPayload `json:"user"`
	IsNew  bool             `json:"is_new"`
	Expire int64            `json:"expire"`
}

// AuthUserPayload mirrors the fields the JS frontend reads (updateUserInfo, etc.).
// All JSON keys are lowercase snake_case so JS can read them directly.
type AuthUserPayload struct {
	ID        uint    `json:"id"`
	Username  string  `json:"username"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Role      string  `json:"role"`
	Balance   float64 `json:"balance"`
	Language  string  `json:"language"`
	Phone     string  `json:"phone"`
	AvatarURL string  `json:"avatar_url"`
}

// validateInitData tries to validate the raw initData string.
// If the first attempt fails, it retries once after url.QueryUnescape to handle
// the double-encoded case that can occur when the frontend sends a string that
// has already been encoded by the browser's fetch layer.
func validateInitData(validator *telegram.Validator, raw string) (initData *telegram.InitData, detail string, err error) {
	// Attempt 1: use the string as-is.
	initData, err = validator.Validate(raw, MaxAuthAge)
	if err == nil {
		return initData, "ok", nil
	}
	detail = err.Error()
	log.Printf("[Auth] Validate(raw) failed: %v", err)

	// Attempt 2: one level of URL unescape.
	unescaped, uerr := url.QueryUnescape(raw)
	if uerr != nil {
		log.Printf("[Auth] url.QueryUnescape also failed: %v", uerr)
		return nil, detail, err
	}
	log.Printf("[Auth] Validate(raw) failed, retrying with QueryUnescape: %v", uerr)

	initData, err = validator.Validate(unescaped, MaxAuthAge)
	if err == nil {
		log.Printf("[Auth] Validate(QueryUnescape) succeeded")
		return initData, "ok (after url.QueryUnescape)", nil
	}
	log.Printf("[Auth] Validate(QueryUnescape) also failed: %v", err)

	// Return the first error as the primary reason.
	return nil, detail, err
}

func AuthTG(db *gorm.DB) gin.HandlerFunc {
	validator := telegram.NewValidator()

	return func(c *gin.Context) {
		var req AuthTGRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[Auth] ShouldBindJSON failed: %v | raw body: %.200q",
				err, c.Request.Body)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "detail": err.Error()})
			return
		}

		// Log what we received (truncated to avoid noise).
		log.Printf("[Auth] POST /api/auth/tg — initData len=%d", len(req.InitData))

		// Validate with fallback for double-encoded input.
		initData, valDetail, valErr := validateInitData(validator, req.InitData)
		if valErr != nil {
			log.Printf("[Auth] Validation failed: primary_error=%v | detail=%s", valErr, valDetail)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":  "invalid initData",
				"detail": valErr.Error(),
			})
			return
		}
		log.Printf("[Auth] Validation ok | user_id=%d username=%s | %s",
			initData.UserID, initData.Username, valDetail)

		var user models.User
		var isNew bool

		result := db.Where("tg_id = ?", initData.UserID).First(&user)
		if result.Error == gorm.ErrRecordNotFound {
			isNew = true

			// Core permission logic: first registered user becomes admin, all others become user.
			var count int64
			db.Model(&models.User{}).Count(&count)

			role := models.RoleUser
			if count == 0 {
				role = models.RoleAdmin
			}

			user = models.User{
				TGID:         initData.UserID,
				Username:     initData.Username,
				FirstName:    initData.FirstName,
				LastName:     initData.LastName,
				Language:     initData.Language,
				Role:         role,
				Balance:      0,
				SessionToken: generateToken(),
			}
			if err := db.Create(&user).Error; err != nil {
				log.Printf("[Auth] db.Create user failed: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
				return
			}
			log.Printf("[Auth] New user created | id=%d username=%s role=%s",
				user.ID, user.Username, user.Role)
		} else if result.Error != nil {
			log.Printf("[Auth] db.First failed: %v", result.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		} else {
			// Existing user — sync profile fields from Telegram.
			log.Printf("[Auth] Existing user | id=%d username=%s — syncing from TG",
				user.ID, user.Username)
			db.Model(&user).Updates(map[string]interface{}{
				"username":   initData.Username,
				"first_name": initData.FirstName,
				"last_name":  initData.LastName,
				"language":   initData.Language,
			})
		}

		newToken := generateToken()
		expireAt := time.Now().Add(TokenTTL)
		if err := db.Model(&user).Update("session_token", newToken).Error; err != nil {
			log.Printf("[Auth] Update session_token failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update token"})
			return
		}

		log.Printf("[Auth] Success | user_id=%d token_prefix=%.10s... expire=%s",
			user.ID, newToken, expireAt.Format(time.RFC3339))

		response := AuthTGResponse{
			Token: newToken,
			User: &AuthUserPayload{
				ID:        user.ID,
				Username:  user.Username,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Role:      user.Role,
				Balance:   user.Balance,
				Language:  user.Language,
				Phone:     "",
				AvatarURL: user.AvatarURL,
			},
			IsNew:  isNew,
			Expire: expireAt.Unix(),
		}
		fmt.Printf("[Debug] Returning to frontend: %v\n", response)
		fmt.Printf("[Final JSON] Sending token: %s\n", newToken)
		c.JSON(http.StatusOK, response)
	}
}

func GetMe(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userVal, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		u := userVal.(*models.User)
		c.JSON(http.StatusOK, AuthUserPayload{
			ID:        u.ID,
			Username:  u.Username,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Role:      u.Role,
			Balance:   u.Balance,
			Language:  u.Language,
			Phone:     "",
			AvatarURL: u.AvatarURL,
		})
	}
}

func generateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
