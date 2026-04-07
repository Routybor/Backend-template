package service

import (
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	cache *JWKSCache
	realm string
}

func NewAuthService(cache *JWKSCache, keycloakURL, realm, clientID string) *AuthService {
	return &AuthService{
		cache: cache,
		realm: realm,
	}
}

func (s *AuthService) ValidateToken(tokenString string) (string, error) {
	kid := ""
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		var ok bool
		kid, ok = token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid not found in token header")
		}

		key, err := s.cache.GetKey(kid)
		if err != nil {
			log.Printf("Failed to get key for kid %s: %v", kid, err)
			return nil, fmt.Errorf("failed to get key: %w", err)
		}
		return key, nil
	}, jwt.WithValidMethods([]string{"RS256"}))

	if err != nil {
		log.Printf("Token validation failed (kid=%s): %v", kid, err)
		return "", fmt.Errorf("invalid token")
	}

	if !token.Valid {
		log.Printf("Token invalid (kid=%s)", kid)
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}

	iss, ok := claims["iss"].(string)
	if !ok || !strings.HasSuffix(iss, fmt.Sprintf("/realms/%s", s.realm)) {
		log.Printf("Invalid issuer: %v", iss)
		return "", fmt.Errorf("invalid issuer")
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("missing subject claim")
	}

	return sub, nil
}

func (s *AuthService) Middleware() gin.HandlerFunc {
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

		sub, err := s.ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}

		c.Request.Header.Set("X-User-ID", sub)
		c.Next()
	}
}
