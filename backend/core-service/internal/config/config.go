package config

import "os"

type Config struct {
	Port     string
	GrpcPort string
}

func Load() *Config {
	return &Config{
		Port:     getEnv("PORT", "8081"),
		GrpcPort: getEnv("GRPC_PORT", "9091"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
