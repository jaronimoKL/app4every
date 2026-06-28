package config

import (
	"os"
)

type Config struct {
	AppEnv     string
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSslMode  string
	RedisHost  string
	RedisPort  string
	JWTSecret  string

	ShikimoriClientID     string
	ShikimoriClientSecret string
	ShikimoriRedirectURI  string

	MasterInviteCode string

	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
}

func LoadConfig() *Config {
	return &Config{
		AppEnv:     getEnv("APP_ENV", "local"),
		Port:       getEnv("PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "superapp_db"),
		DBSslMode:  getEnv("DB_SSLMODE", "disable"),
		RedisHost:  getEnv("REDIS_HOST", "localhost"),
		RedisPort:  getEnv("REDIS_PORT", "6379"),
		JWTSecret:  getEnv("JWT_SECRET", "default-jwt-secret-key-change-me"),
		ShikimoriClientID:     getEnv("SHIKIMORI_CLIENT_ID", ""),
		ShikimoriClientSecret: getEnv("SHIKIMORI_CLIENT_SECRET", ""),
		ShikimoriRedirectURI:  getEnv("SHIKIMORI_REDIRECT_URI", "http://localhost/api/v1/auth/shikimori/callback"),
		MasterInviteCode:      getEnv("MASTER_INVITE_CODE", ""),
		SMTPHost:              getEnv("SMTP_HOST", ""),
		SMTPPort:              getEnv("SMTP_PORT", "465"),
		SMTPUser:              getEnv("SMTP_USER", ""),
		SMTPPassword:          getEnv("SMTP_PASSWORD", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
