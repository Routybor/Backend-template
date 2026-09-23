package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func String(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Int(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func Duration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func Required(keys ...string) error {
	for _, k := range keys {
		if os.Getenv(k) == "" {
			return fmt.Errorf("missing required env %s", k)
		}
	}
	return nil
}
