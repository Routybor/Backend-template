package auth

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

type JWKSProvider struct {
	mu         sync.RWMutex
	keys       map[string]*rsa.PublicKey
	fetchedAt  time.Time
	ttl        time.Duration
	minRefetch time.Duration
	url        string
	client     *http.Client
}

func NewJWKSProvider(keycloakURL, realm string) *JWKSProvider {
	return &JWKSProvider{
		keys:       make(map[string]*rsa.PublicKey),
		ttl:        15 * time.Minute,
		minRefetch: time.Minute,
		url:        fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", keycloakURL, realm),
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *JWKSProvider) Key(kid string) (*rsa.PublicKey, error) {
	p.mu.RLock()
	key, ok := p.keys[kid]
	fresh := time.Since(p.fetchedAt) < p.ttl
	p.mu.RUnlock()
	if ok && fresh {
		return key, nil
	}

	if err := p.refresh(); err != nil {
		return nil, err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()
	if key, ok := p.keys[kid]; ok {
		return key, nil
	}
	return nil, fmt.Errorf("key not found: %s", kid)
}

func (p *JWKSProvider) refresh() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if time.Since(p.fetchedAt) < p.minRefetch {
		return nil
	}

	resp, err := p.client.Get(p.url)
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

	for _, k := range jwks.Keys {
		if k.Kty != "RSA" || k.Use != "sig" {
			continue
		}
		n, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		e, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		p.keys[k.Kid] = &rsa.PublicKey{
			N: new(big.Int).SetBytes(n),
			E: int(new(big.Int).SetBytes(e).Int64()),
		}
	}

	p.fetchedAt = time.Now()
	return nil
}
