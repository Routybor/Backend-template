package handler

import (
	"net/http"

	"core-service/internal/dto"
	"core-service/internal/service"

	"github.com/gin-gonic/gin"
)

type ItemHandler struct {
	service *service.ItemService
}

func NewItemHandler(svc *service.ItemService) *ItemHandler {
	return &ItemHandler{service: svc}
}

func (h *ItemHandler) GetAll(c *gin.Context) {
	items := h.service.GetAllItems()
	c.JSON(http.StatusOK, dto.ListResponse[dto.Item]{
		Items: items,
		Total: len(items),
	})
}

func (h *ItemHandler) Create(c *gin.Context) {
	var req dto.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	item := h.service.CreateItem(req)
	c.JSON(http.StatusCreated, item)
}

func (h *ItemHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	item, ok := h.service.GetItem(id)
	if !ok {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "item not found",
		})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *ItemHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if !h.service.DeleteItem(id) {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "item not found",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(c *gin.Context) {
	c.JSON(http.StatusOK, dto.HealthResponse{Status: "healthy"})
}
