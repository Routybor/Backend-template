package auth

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Validator struct {
	jwks  *JWKSProvider
	realm string
}

func NewValidator(keycloakURL, realm string) *Validator {
	return &Validator{
		jwks:  NewJWKSProvider(keycloakURL, realm),
		realm: realm,
	}
}

func (v *Validator) Validate(tokenString string) (string, error) {
	var kid string

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		var ok bool
		kid, ok = t.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid not found in token header")
		}

		return v.jwks.Key(kid)
	}, jwt.WithValidMethods([]string{"RS256"}))

	if err != nil || !token.Valid {
		slog.Warn("token validation failed", "kid", kid, "error", err)
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}

	iss, ok := claims["iss"].(string)
	if !ok || !strings.HasSuffix(iss, fmt.Sprintf("/realms/%s", v.realm)) {
		return "", fmt.Errorf("invalid issuer")
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("missing subject claim")
	}

	return sub, nil
}
