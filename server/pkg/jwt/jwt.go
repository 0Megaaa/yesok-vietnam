package jwt

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"yesok-vietnam/server/models"
)

var (
	ErrTokenExpired = errors.New("token has expired")
	ErrTokenInvalid = errors.New("token is invalid")
)

type Claims struct {
	UID     uint   `json:"uid"`
	Role    string `json:"role"`
	IsAdmin bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// Config holds the signing key and token TTL. In production this should be
// injected from config.Global.JWT.Secret (set via JWT_SECRET env var).
var Secret = []byte(getEnv("JWT_SECRET", "/MIGdDnsUBlrxKZxW4loyBWkJCL0yJ6fi9MPZIhuWdqiui5XnpYhRmFWfuP6lnvf"))

// TokenTTL is for C-end users (7 days, unchanged for backward compatibility).
var TokenTTL = 7 * 24 * time.Hour

// AdminTokenTTL is for B-end admins, configurable via ADMIN_JWT_TTL env var (default 8 hours).
var AdminTokenTTL = getDurationEnv("ADMIN_JWT_TTL", 8*time.Hour)

// getDurationEnv parses a duration string from env var, returns fallback on error.
func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil && d > 0 {
			return d
		}
	}
	return fallback
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// TokenHash returns the SHA256 hex digest of a raw token string.
func TokenHash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// SignWithTTL creates a JWT with a custom TTL.
func SignWithTTL(uid uint, role string, ttl time.Duration) (string, int64, error) {
	expireAt := time.Now().Add(ttl)
	claims := Claims{
		UID:     uid,
		Role:    role,
		IsAdmin: role == models.RoleAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(Secret)
	return signed, expireAt.Unix(), err
}

// SignAdmin creates a JWT for B-end admin using AdminTokenTTL.
func SignAdmin(uid uint, role string) (string, int64, error) {
	return SignWithTTL(uid, role, AdminTokenTTL)
}

// Sign creates a JWT for C-end users using TokenTTL (backward compatible).
func Sign(uid uint, role string) (string, int64, error) {
	return SignWithTTL(uid, role, TokenTTL)
}

func Validate(raw string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return Secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}
