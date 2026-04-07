package main

import (
	"context"
	"gateway/internal/config"
	"gateway/internal/handler"
	"gateway/internal/middleware"
	"gateway/internal/service"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	slog.Info("starting gateway")

	cfg := config.Load()

	proxy, err := service.NewReverseProxy(cfg.CoreServiceURL)
	if err != nil {
		slog.Error("failed to create proxy", "error", err)
		os.Exit(1)
	}

	jwksCache := service.NewJWKSCache(cfg.KeycloakURL, cfg.Realm)
	authService := service.NewAuthService(jwksCache, cfg.KeycloakURL, cfg.Realm, cfg.ClientID)

	healthHandler := handler.NewHealthHandler()
	proxyHandler := handler.NewProxyHandler(proxy)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger())
	router.Use(securityHeaders())
	router.Use(middleware.GzipMiddleware())
	router.Use(middleware.RateLimitMiddleware(cfg.RateLimitRequests, time.Duration(cfg.RateLimitWindowSec)*time.Second))

	router.GET("/health", healthHandler.Check)
	router.Use(authService.Middleware())
	router.Use(middleware.CircuitBreakerMiddleware(cfg.CircuitThreshold, time.Duration(cfg.CircuitTimeoutSec)*time.Second))

	router.Any("/items", proxyHandler.HandleItems)
	router.Any("/items/*path", proxyHandler.HandleItemsWildcard)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("gateway listening", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		slog.Info("request completed",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency", time.Since(start),
			"client_ip", c.ClientIP(),
		)
	}
}

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}
