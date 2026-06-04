package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/shashisharma307703/vedantam/config"
	"github.com/shashisharma307703/vedantam/internal/domain/auth"
	"github.com/shashisharma307703/vedantam/internal/repository"
)

// OIDCDiscoveryService handles OIDC discovery document fetching and caching
type OIDCDiscoveryService interface {
	// GetDiscoveryDocument fetches and caches the OIDC discovery document
	GetDiscoveryDocument(ctx context.Context, discoveryURL string) (*auth.OIDCDiscoveryDocument, error)
}

type oidcDiscoveryService struct {
	config     *config.AuthConfig
	cacheRepo  repository.OIDCDiscoveryCacheRepository
	httpClient *http.Client
}

// NewOIDCDiscoveryService creates a new OIDC discovery service
func NewOIDCDiscoveryService(
	cfg *config.AuthConfig,
	cacheRepo repository.OIDCDiscoveryCacheRepository,
) OIDCDiscoveryService {
	return &oidcDiscoveryService{
		config:    cfg,
		cacheRepo: cacheRepo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetDiscoveryDocument fetches and caches the OIDC discovery document
func (s *oidcDiscoveryService) GetDiscoveryDocument(ctx context.Context, discoveryURL string) (*auth.OIDCDiscoveryDocument, error) {
	// Extract issuer from discovery URL (remove /.well-known/openid-configuration)
	issuer := s.extractIssuer(discoveryURL)

	// Try to get from cache first
	cached, err := s.cacheRepo.GetByIssuer(ctx, issuer)
	if err == nil && !cached.IsExpired() {
		return &cached.Document, nil
	}

	// Fetch from provider
	doc, err := s.fetchDiscoveryDocument(ctx, discoveryURL)
	if err != nil {
		return nil, err
	}

	// Cache the result
	cacheEntry := &auth.OIDCDiscoveryCache{
		Issuer:    issuer,
		Document:  *doc,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(s.config.OIDCDiscoveryCacheTTL),
	}

	// Try to create cache entry (ignore errors if it already exists)
	_, _ = s.cacheRepo.Create(ctx, cacheEntry)

	return doc, nil
}

// fetchDiscoveryDocument makes HTTP request to fetch the discovery document
func (s *oidcDiscoveryService) fetchDiscoveryDocument(ctx context.Context, discoveryURL string) (*auth.OIDCDiscoveryDocument, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request", auth.ErrOIDCDiscoveryFailed)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch discovery document", auth.ErrOIDCDiscoveryFailed)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: HTTP %d: %s", auth.ErrOIDCDiscoveryFailed, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read response body", auth.ErrOIDCDiscoveryFailed)
	}

	var doc auth.OIDCDiscoveryDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("%w: invalid JSON response", auth.ErrOIDCDiscoveryFailed)
	}

	return &doc, nil
}

// extractIssuer extracts the issuer URL from the discovery URL
// Discovery URL format: https://issuer.com/.well-known/openid-configuration
// Returns: https://issuer.com
func (s *oidcDiscoveryService) extractIssuer(discoveryURL string) string {
	// Remove /.well-known/openid-configuration suffix
	issuer := discoveryURL
	suffix := "/.well-known/openid-configuration"
	if len(issuer) > len(suffix) && issuer[len(issuer)-len(suffix):] == suffix {
		issuer = issuer[:len(issuer)-len(suffix)]
	}
	return issuer
}
