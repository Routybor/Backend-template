package service

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"
)

type JWKSCache struct {
	keys        map[string]*rsa.PublicKey
	mu          sync.RWMutex
	lastFetch   time.Time
	ttl         time.Duration
	keycloakURL string
	realm       string
}

type jwkSet struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func NewJWKSCache(keycloakURL, realm string) *JWKSCache {
	return &JWKSCache{
		keys:        make(map[string]*rsa.PublicKey),
		ttl:         15 * time.Minute,
		keycloakURL: keycloakURL,
		realm:       realm,
	}
}

func (c *JWKSCache) GetKey(kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	if key, ok := c.keys[kid]; ok && time.Since(c.lastFetch) < c.ttl {
		c.mu.RUnlock()
		return key, nil
	}
	c.mu.RUnlock()

	if err := c.refresh(); err != nil {
		return nil, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	if key, ok := c.keys[kid]; ok {
		return key, nil
	}
	return nil, fmt.Errorf("key not found: %s", kid)
}

func (c *JWKSCache) refresh() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Since(c.lastFetch) < time.Minute {
		return nil
	}

	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", c.keycloakURL, c.realm)
	resp, err := http.Get(jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status: %d", resp.StatusCode)
	}

	var jwks jwkSet
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	for _, key := range jwks.Keys {
		if key.Kty != "RSA" || key.Use != "sig" {
			continue
		}

		n, err := base64.RawURLEncoding.DecodeString(key.N)
		if err != nil {
			continue
		}
		e, err := base64.RawURLEncoding.DecodeString(key.E)
		if err != nil {
			continue
		}

		pubKey := &rsa.PublicKey{
			N: new(big.Int).SetBytes(n),
			E: int(new(big.Int).SetBytes(e).Int64()),
		}
		c.keys[key.Kid] = pubKey
	}

	c.lastFetch = time.Now()
	return nil
}
