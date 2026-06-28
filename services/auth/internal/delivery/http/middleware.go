package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"app4every/services/auth/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "userID"

func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			var tokenString string
			
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString = parts[1]
				}
			}

			// Если токена нет в заголовке, проверяем query параметры (для OAuth и WebSocket)
			if tokenString == "" {
				tokenString = r.URL.Query().Get("token")
			}
			if tokenString == "" {
				// Shikimori возвращает token в параметре state
				tokenString = r.URL.Query().Get("state")
			}

			if tokenString == "" {
				http.Error(w, `{"error":"unauthorized","message":"missing token"}`, http.StatusUnauthorized)
				return
			}


			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(cfg.JWTSecret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, `{"error":"unauthorized","message":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, `{"error":"unauthorized","message":"invalid claims"}`, http.StatusUnauthorized)
				return
			}

			subFloat, ok := claims["sub"].(float64)
			if !ok {
				http.Error(w, `{"error":"unauthorized","message":"invalid subject claim"}`, http.StatusUnauthorized)
				return
			}

			userID := int64(subFloat)
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
