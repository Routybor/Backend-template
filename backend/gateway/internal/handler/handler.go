package handler

import (
	"net/http"

	"gateway/internal/service"

	"github.com/gin-gonic/gin"
)

type ProxyHandler struct {
	proxy *service.ReverseProxy
}

func NewProxyHandler(proxy *service.ReverseProxy) *ProxyHandler {
	return &ProxyHandler{proxy: proxy}
}

func (h *ProxyHandler) HandleItems(c *gin.Context) {
	h.proxy.ServeHTTP(c.Writer, c.Request)
}

func (h *ProxyHandler) HandleItemsWildcard(c *gin.Context) {
	h.proxy.ServeHTTP(c.Writer, c.Request)
}

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}
