package config

import "os"

type Config struct {
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSslMode  string
	JWTSecret  string
	RedisHost      string
	RedisPort      string
	AuthServiceURL string
}

func LoadConfig() *Config {
	return &Config{
		Port:       getEnv("PORT", "8082"),
		DBHost:     getEnv("DB_HOST", "postgres"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "super_user"),
		DBPassword: getEnv("DB_PASSWORD", "super_password"),
		DBName:     getEnv("DB_NAME", "app4every_db"),
		DBSslMode:  getEnv("DB_SSLMODE", "disable"),
		JWTSecret:  getEnv("JWT_SECRET", "change_me_in_production"),
		RedisHost:  getEnv("REDIS_HOST", "redis"),
		RedisPort:  getEnv("REDIS_PORT", "6379"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://app4every-auth:8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
