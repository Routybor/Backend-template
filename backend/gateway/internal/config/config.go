package config

import "os"

type Config struct {
	Port               string
	CoreServiceURL     string
	KeycloakURL        string
	Realm              string
	ClientID           string
	RateLimitRequests  int
	RateLimitWindowSec int
	CircuitThreshold   int
	CircuitTimeoutSec  int
}

func Load() *Config {
	return &Config{
		Port:               getEnv("GATEWAY_PORT", "8080"),
		CoreServiceURL:     getEnv("CORE_SERVICE_URL", "http://localhost:8081"),
		KeycloakURL:        getEnv("KEYCLOAK_URL", "http://localhost:8180"),
		Realm:              getEnv("KEYCLOAK_REALM", "master"),
		ClientID:           getEnv("KEYCLOAK_CLIENT_ID", "gateway"),
		RateLimitRequests:  getEnvInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindowSec: getEnvInt("RATE_LIMIT_WINDOW_SEC", 60),
		CircuitThreshold:   getEnvInt("CIRCUIT_THRESHOLD", 5),
		CircuitTimeoutSec:  getEnvInt("CIRCUIT_TIMEOUT_SEC", 30),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		for _, c := range value {
			if c >= '0' && c <= '9' {
				result = result*10 + int(c-'0')
			} else {
				return defaultValue
			}
		}
		return result
	}
	return defaultValue
}
