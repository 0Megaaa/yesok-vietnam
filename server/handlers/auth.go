package handlers

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"yesok-vietnam/server/models"
	"yesok-vietnam/server/pkg/jwt"
	"yesok-vietnam/server/pkg/telegram"
)

const MaxAuthAge = 24 * time.Hour

// adminPasswords maps admin usernames to bcrypt-hashed passwords.
// Default: username "admin", password "admin123" (hash for "admin123").
// To change the password: generate a new hash with bcrypt.GenerateFromPassword.
// To add more admins: duplicate the entry with a different username:hash pair.
var adminPasswords = map[string]string{
	"admin": "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
}

// ─── Shared request types ─────────────────────────────────────────────────────

type AuthTGRequest struct {
	InitData string `json:"initData" binding:"required"`
}

type AuthAdminRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ─── Client auth: POST /api/v1/client/auth/tg ─────────────────────────────────

func AuthTG(db *gorm.DB) gin.HandlerFunc {
	validator := telegram.NewValidator()

	return func(c *gin.Context) {
		var req AuthTGRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[AuthTG] ShouldBindJSON failed: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "detail": err.Error()})
			return
		}

		log.Printf("[AuthTG] POST /api/v1/client/auth/tg — initData len=%d", len(req.InitData))

		initData, detail, valErr := validateInitData(validator, req.InitData)
		if valErr != nil {
			log.Printf("[AuthTG] validation failed: %v | detail=%s", valErr, detail)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid initData", "detail": valErr.Error()})
			return
		}
		log.Printf("[AuthTG] ok | user_id=%d username=%s | %s", initData.UserID, initData.Username, detail)

		var user models.User
		var isNew bool

		result := db.Where("tg_id = ?", initData.UserID).First(&user)
		if result.Error == gorm.ErrRecordNotFound {
			isNew = true
			var count int64
			db.Model(&models.User{}).Count(&count)

			role := models.RoleUser
			if count == 0 {
				role = models.RoleAdmin // first registered user becomes admin
			}

			user = models.User{
				TGID:      initData.UserID,
				Username:  initData.Username,
				FirstName: initData.FirstName,
				LastName:  initData.LastName,
				Language:  initData.Language,
				Role:      role,
				Balance:   0,
			}
			if err := db.Create(&user).Error; err != nil {
				log.Printf("[AuthTG] db.Create failed: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
				return
			}
			log.Printf("[AuthTG] new user created | id=%d username=%s role=%s", user.ID, user.Username, user.Role)
		} else if result.Error != nil {
			log.Printf("[AuthTG] db.First failed: %v", result.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		} else {
			// Sync profile from Telegram on every login.
			log.Printf("[AuthTG] existing user | id=%d username=%s — syncing from TG", user.ID, user.Username)
			db.Model(&user).Updates(map[string]interface{}{
				"username":   initData.Username,
				"first_name": initData.FirstName,
				"last_name":  initData.LastName,
				"language":   initData.Language,
			})
		}

		// Issue JWT.
		jwtToken, expireUnix, err := jwt.Sign(user.ID, user.Role)
		if err != nil {
			log.Printf("[AuthTG] jwt.Sign failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
			return
		}

		// Store JWT in DB session_token column (for multi-login tracking / force-logout).
		if err := db.Model(&user).Update("session_token", jwtToken).Error; err != nil {
			log.Printf("[AuthTG] update session_token warning: %v (non-fatal)", err)
		}

		log.Printf("[AuthTG] success | user_id=%d role=%s token_prefix=%.10s... expire=%d",
			user.ID, user.Role, jwtToken, expireUnix)

		c.JSON(http.StatusOK, gin.H{
			"token":  jwtToken,
			"user":   userPayload(&user),
			"is_new": isNew,
			"expire": expireUnix,
		})
	}
}

// validateInitData tries to validate the raw initData string.
// If the first attempt fails, it retries once after url.QueryUnescape to handle
// the double-encoded case that can occur when the frontend sends a string that
// has already been encoded by the browser's fetch layer.
func validateInitData(validator *telegram.Validator, raw string) (*telegram.InitData, string, error) {
	initData, err := validator.Validate(raw, MaxAuthAge)
	if err == nil {
		return initData, "ok", nil
	}
	detail := err.Error()
	log.Printf("[AuthTG] Validate(raw) failed: %v", err)

	unescaped, uerr := url.QueryUnescape(raw)
	if uerr != nil {
		log.Printf("[AuthTG] url.QueryUnescape also failed: %v", uerr)
		return nil, detail, err
	}
	log.Printf("[AuthTG] Validate(raw) failed, retrying with QueryUnescape")

	initData, err = validator.Validate(unescaped, MaxAuthAge)
	if err == nil {
		log.Printf("[AuthTG] Validate(QueryUnescape) succeeded")
		return initData, "ok (after url.QueryUnescape)", nil
	}
	log.Printf("[AuthTG] Validate(QueryUnescape) also failed: %v", err)
	return nil, detail, err
}

// ─── Admin auth: POST /api/v1/admin/auth/login ───────────────────────────────

func AuthAdmin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AuthAdminRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[AuthAdmin] ShouldBindJSON failed: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "detail": err.Error()})
			return
		}

		log.Printf("[AuthAdmin] POST /api/v1/admin/auth/login — username=%s", req.Username)

		// Look up bcrypt hash.
		storedHash, ok := adminPasswords[req.Username]
		if !ok {
			log.Printf("[AuthAdmin] unknown username: %s", req.Username)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)); err != nil {
			log.Printf("[AuthAdmin] wrong password for user: %s", req.Username)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		// Admins have UID=0; the JWT carries isAdmin=true via Role=admin.
		jwtToken, expireUnix, err := jwt.Sign(0, models.RoleAdmin)
		if err != nil {
			log.Printf("[AuthAdmin] jwt.Sign failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
			return
		}

		log.Printf("[AuthAdmin] success | username=%s token_prefix=%.10s...", req.Username, jwtToken)

		c.JSON(http.StatusOK, gin.H{
			"token": jwtToken,
			"user": gin.H{
				"id":       0,
				"username": req.Username,
				"role":     models.RoleAdmin,
				"is_admin": true,
			},
			"expire": expireUnix,
		})
	}
}

// ─── Admin auth: POST /api/v1/admin/auth/logout ──────────────────────────────

func AuthLogout() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, _ := c.Get("uid")
		role, _ := c.Get("role")
		log.Printf("[AuthLogout] uid=%v role=%v", uid, role)
		c.JSON(http.StatusOK, gin.H{"message": "logged out"})
	}
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// userPayload converts a DB User model to a JSON-safe map.
func userPayload(u *models.User) gin.H {
	return gin.H{
		"id":         u.ID,
		"username":   u.Username,
		"first_name": u.FirstName,
		"last_name":  u.LastName,
		"role":       u.Role,
		"balance":    u.Balance,
		"language":   u.Language,
		"phone":      "",
		"avatar_url": u.AvatarURL,
	}
}
