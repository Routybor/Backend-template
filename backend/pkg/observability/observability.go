package observability

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func HealthHandler(checks map[string]func(ctx context.Context) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := make(map[string]string, len(checks))
		healthy := true

		for name, check := range checks {
			if err := check(r.Context()); err != nil {
				results[name] = "down"
				healthy = false
				slog.Error("health check failed", "dependency", name, "error", err)
			} else {
				results[name] = "up"
			}
		}

		status := "healthy"
		code := http.StatusOK
		if !healthy {
			status = "unhealthy"
			code = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]any{"status": status, "checks": results})
	}
}
