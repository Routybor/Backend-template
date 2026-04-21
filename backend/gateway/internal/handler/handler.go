package handler

import (
	"net/http"

	"gateway/internal/service"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProxyHandler struct {
	client *service.ItemGrpcClient
}

func NewProxyHandler(client *service.ItemGrpcClient) *ProxyHandler {
	return &ProxyHandler{client: client}
}

func (h *ProxyHandler) HandleItems(c *gin.Context) {
	switch c.Request.Method {
	case http.MethodGet:
		items, err := h.client.GetAll(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": items.Items, "total": items.Total})
	case http.MethodPost:
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
			return
		}
		item, err := h.client.Create(c.Request.Context(), req.Name, req.Description)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, item)
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method_not_allowed"})
	}
}

func (h *ProxyHandler) HandleItemsWildcard(c *gin.Context) {
	id := c.Param("path")
	if id == "" {
		h.HandleItems(c)
		return
	}

	switch c.Request.Method {
	case http.MethodGet:
		item, err := h.client.GetByID(c.Request.Context(), id)
		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "item not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, item)
	case http.MethodDelete:
		_, err := h.client.Delete(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method_not_allowed"})
	}
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
