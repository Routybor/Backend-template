package config

import (
	"fmt"

	pkgconfig "backend/pkg/config"
)

type Config struct {
	Port               string
	CoreServiceGrpc    string
	KeycloakURL        string
	Realm              string
	RateLimitRequests  int
	RateLimitWindowSec int
	CircuitThreshold   int
	CircuitTimeoutSec  int
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:               pkgconfig.String("GATEWAY_PORT", "8080"),
		CoreServiceGrpc:    pkgconfig.String("CORE_SERVICE_GRPC", "localhost:9091"),
		KeycloakURL:        pkgconfig.String("KEYCLOAK_URL", "http://localhost:8180"),
		Realm:              pkgconfig.String("KEYCLOAK_REALM", "master"),
		RateLimitRequests:  pkgconfig.Int("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindowSec: pkgconfig.Int("RATE_LIMIT_WINDOW_SEC", 60),
		CircuitThreshold:   pkgconfig.Int("CIRCUIT_THRESHOLD", 5),
		CircuitTimeoutSec:  pkgconfig.Int("CIRCUIT_TIMEOUT_SEC", 30),
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	switch {
	case c.RateLimitRequests <= 0:
		return fmt.Errorf("RATE_LIMIT_REQUESTS must be positive")
	case c.RateLimitWindowSec <= 0:
		return fmt.Errorf("RATE_LIMIT_WINDOW_SEC must be positive")
	case c.CircuitThreshold <= 0:
		return fmt.Errorf("CIRCUIT_THRESHOLD must be positive")
	case c.CircuitTimeoutSec <= 0:
		return fmt.Errorf("CIRCUIT_TIMEOUT_SEC must be positive")
	}
	return nil
}
