package handlers

import (
	"net/http"

	"gateway/internal/dto"

	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, dto.HealthResponse{Status: "healthy"})
}
