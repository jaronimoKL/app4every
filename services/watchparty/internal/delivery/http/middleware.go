package http

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   int64
	Username string
}

func ValidateToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	sub, ok := mapClaims["sub"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid sub")
	}

	return &Claims{
		UserID:   int64(sub),
		Username: fmt.Sprintf("User %d", int64(sub)),
	}, nil
}
