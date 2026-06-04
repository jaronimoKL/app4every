package http

import (
	"context"
	"net/http"
	"strings"

	"app4every/services/notebook/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

// contextKey — тип для ключей контекста, предотвращает коллизии.
type contextKey string

const UserIDKey contextKey = "userID"

// AuthMiddleware проверяет JWT-токен из заголовка Authorization.
// Notebook-сервис не обращается к Redis — JWT верифицируется локально
// той же подписью что и в auth-сервисе (одинаковый JWT_SECRET).
func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error":"unauthorized","message":"missing token"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
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

			// sub хранится как float64 в JSON — конвертируем в int64
			sub, ok := claims["sub"].(float64)
			if !ok {
				http.Error(w, `{"error":"unauthorized","message":"invalid subject"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, int64(sub))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
