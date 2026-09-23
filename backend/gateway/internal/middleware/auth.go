package middleware

import (
	"strings"

	"backend/pkg/auth"

	"github.com/gin-gonic/gin"
)

const userIDHeader = "X-User-ID"

type Auth struct {
	validator *auth.Validator
}

func NewAuth(validator *auth.Validator) *Auth {
	return &Auth{validator: validator}
}

func (a *Auth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid authorization header format"})
			return
		}

		sub, err := a.validator.Validate(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}

		c.Request.Header.Set(userIDHeader, sub)
		c.Next()
	}
}

func UserID(c *gin.Context) (string, bool) {
	id := c.GetHeader(userIDHeader)
	return id, id != ""
}
