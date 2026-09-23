package handlers

import (
	"log/slog"
	"net/http"

	"backend/pkg/requestid"

	"gateway/internal/dto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func respondError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		slog.Error("items call failed", "error", err, "request_id", requestid.FromContext(c.Request.Context()))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error"})
		return
	}

	switch st.Code() {
	case codes.NotFound:
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: st.Message()})
	case codes.InvalidArgument:
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_request", Message: st.Message()})
	case codes.Unavailable:
		slog.Error("core-service unavailable", "request_id", requestid.FromContext(c.Request.Context()))
		c.JSON(http.StatusBadGateway, dto.ErrorResponse{Error: "service_unavailable"})
	case codes.DeadlineExceeded:
		c.JSON(http.StatusGatewayTimeout, dto.ErrorResponse{Error: "gateway_timeout"})
	default:
		slog.Error("items call failed", "code", st.Code(), "message", st.Message(), "request_id", requestid.FromContext(c.Request.Context()))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error"})
	}
}
