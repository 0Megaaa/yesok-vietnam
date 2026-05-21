package jwt

import (
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
var Secret = []byte(getEnv("JWT_SECRET", "yesok-vietnam-jwt-secret-change-in-production"))
var TokenTTL = 7 * 24 * time.Hour // 7 days

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func Sign(uid uint, role string) (string, int64, error) {
	expireAt := time.Now().Add(TokenTTL)
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
