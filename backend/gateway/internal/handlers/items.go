package handlers

import (
	"net/http"

	"gateway/internal/service"

	"github.com/gin-gonic/gin"
)

type ItemsHandler struct {
	client *service.ItemsClient
}

func NewItemsHandler(client *service.ItemsClient) *ItemsHandler {
	return &ItemsHandler{client: client}
}

func (h *ItemsHandler) HandleItems(c *gin.Context) {
	switch c.Request.Method {
	case http.MethodGet:
		items, err := h.client.GetAll(c.Request.Context())
		if err != nil {
			respondError(c, err)
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
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, item)
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method_not_allowed"})
	}
}

func (h *ItemsHandler) HandleItemsWildcard(c *gin.Context) {
	id := c.Param("path")
	if id == "" {
		h.HandleItems(c)
		return
	}

	switch c.Request.Method {
	case http.MethodGet:
		item, err := h.client.GetByID(c.Request.Context(), id)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, item)
	case http.MethodDelete:
		_, err := h.client.Delete(c.Request.Context(), id)
		if err != nil {
			respondError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method_not_allowed"})
	}
}
