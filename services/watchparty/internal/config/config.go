package config

import (
	"log"
	"os"
)

type Config struct {
	Port           string
	JWTSecret      string
	AuthServiceURL string
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Println("WARNING: JWT_SECRET is not set, using default for development only")
		jwtSecret = "your-very-secure-local-jwt-secret-key"
	}

	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://auth-service:8080"
	}

	return &Config{
		Port:           port,
		JWTSecret:      jwtSecret,
		AuthServiceURL: authServiceURL,
	}
}
