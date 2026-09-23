package main

import (
	"log/slog"
	"os"

	"gateway/internal/app"
	"gateway/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}

	if err := app.New(cfg).Run(); err != nil {
		slog.Error("gateway exited", "error", err)
		os.Exit(1)
	}
}
