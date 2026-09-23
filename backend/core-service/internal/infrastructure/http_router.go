package infrastructure

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	pkgconfig "backend/pkg/config"
	"backend/pkg/observability"

	"core-service/internal/domain"
	"core-service/internal/service"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Port     string
	GrpcPort string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:     pkgconfig.String("PORT", "8081"),
		GrpcPort: pkgconfig.String("GRPC_PORT", "9091"),
	}
	for name, value := range map[string]string{"PORT": cfg.Port, "GRPC_PORT": cfg.GrpcPort} {
		if _, err := strconv.Atoi(value); err != nil {
			return nil, fmt.Errorf("%s must be numeric, got %q", name, value)
		}
	}
	return cfg, nil
}

type itemResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

type listResponse struct {
	Items []itemResponse `json:"items"`
	Total int            `json:"total"`
}

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func NewHTTPRouter(svc *service.ItemService, checks map[string]func(context.Context) error) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery(), requestLogger(), securityHeaders())

	health := gin.WrapH(observability.HealthHandler(checks))
	router.GET("/health", health)
	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	router.GET("/health/ready", health)
	router.GET("/metrics", gin.WrapH(observability.MetricsHandler()))

	h := &httpHandler{svc: svc}
	router.GET("/items", h.GetAll)
	router.POST("/items", h.Create)
	router.GET("/items/:id", h.GetByID)
	router.DELETE("/items/:id", h.Delete)

	return router
}

type httpHandler struct {
	svc *service.ItemService
}

func (h *httpHandler) GetAll(c *gin.Context) {
	items := h.svc.GetAll(c.Request.Context())
	out := make([]itemResponse, len(items))
	for i, item := range items {
		out[i] = toItemResponse(item)
	}
	c.JSON(http.StatusOK, listResponse{Items: out, Total: len(out)})
}

func (h *httpHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required,min=1,max=255"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	item, err := h.svc.Create(c.Request.Context(), req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_input"})
		return
	}
	c.JSON(http.StatusCreated, toItemResponse(item))
}

func (h *httpHandler) GetByID(c *gin.Context) {
	item, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{Error: "not_found", Message: "item not found"})
		return
	}
	c.JSON(http.StatusOK, toItemResponse(item))
}

func (h *httpHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, errorResponse{Error: "not_found", Message: "item not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func toItemResponse(item domain.Item) itemResponse {
	return itemResponse{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		CreatedAt:   item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.Format(time.RFC3339),
	}
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
			"user_id", c.GetHeader("X-User-ID"),
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
