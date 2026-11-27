package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	Access  = "access"
	Refresh = "refresh"
)

type TokenManager struct {
	secret     []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func NewTokenManager(secret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{
		secret:     []byte(secret),
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
	}
}

func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil // بدون padding (=)
}

func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (t *TokenManager) GenerateAccessToken(userID int, email, name string) (string, time.Time, error) {
	jti, err := randomString(16)
	if err != nil {
		return "", time.Time{}, err
	}
	exp := time.Now().UTC().Add(t.AccessTTL)
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"name":    name,
		"jti":     jti,
		"exp":     exp.Unix(),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(t.secret)
	return signed, exp, err
}

func (t *TokenManager) GenerateRefreshToken() (string, time.Time, error) {
	raw, err := randomString(32)
	if err != nil {
		return "", time.Time{}, err
	}
	return raw, time.Now().UTC().Add(t.RefreshTTL), nil
}

func (t *TokenManager) Parse(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(*jwt.Token) (any, error) {
		return t.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}
