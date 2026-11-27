package middleware

import (
	"auth-api/pkg/auth"
	"context"
	"net/http"
	"strings"
)

type ctxKey string

const userKey ctxKey = "user"

type TokenChecker interface {
	IsValid(tokenType, hash string) (bool, error)
}
type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

func Auth(tokens *auth.TokenManager, store TokenChecker) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				unauthorized(w)
				return
			}
			raw := strings.TrimPrefix(header, "Bearer ")
			claims, err := tokens.Parse(raw)
			if err != nil {
				unauthorized(w)
				return
			}

			ok, err := store.IsValid(auth.Access, auth.HashToken(raw))
			if err != nil || !ok {
				unauthorized(w)
				return
			}

			userID, _ := claims["user_id"].(float64)
			email, _ := claims["email"].(string)
			name, _ := claims["name"].(string)

			ctx := context.WithValue(r.Context(), userKey, User{ID: int(userID), Email: email, Name: name})
			next(w, r.WithContext(ctx))
		}
	}
}

func UserFrom(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userKey).(User)
	return u, ok
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"unauthorized"}`))
}
