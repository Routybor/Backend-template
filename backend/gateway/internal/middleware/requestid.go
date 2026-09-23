package middleware

import (
	"backend/pkg/requestid"

	"github.com/gin-gonic/gin"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(requestid.Header)
		if id == "" {
			id = requestid.New()
		}
		c.Header(requestid.Header, id)
		c.Request = c.Request.WithContext(requestid.WithContext(c.Request.Context(), id))
		c.Next()
	}
}
