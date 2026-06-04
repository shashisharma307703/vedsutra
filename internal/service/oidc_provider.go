package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shashisharma307703/vedantam/internal/domain/auth"
)

// OIDCProviderImpl implements the OIDCProvider interface for Keycloak/generic OIDC
type OIDCProviderImpl struct {
	name         string
	providerType auth.ProviderType
	config       map[string]interface{}
	discoveryDoc *auth.OIDCDiscoveryDocument
	discoveryService OIDCDiscoveryService
	httpClient   *http.Client
}

// NewOIDCProvider creates a new OIDC provider instance
func NewOIDCProvider(
	name string,
	providerType auth.ProviderType,
	config map[string]interface{},
	discoveryService OIDCDiscoveryService,
) *OIDCProviderImpl {
	return &OIDCProviderImpl{
		name:                 name,
		providerType:         providerType,
		config:               config,
		discoveryService:     discoveryService,
		httpClient:           &http.Client{Timeout: 10 * time.Second},
	}
}

// Authenticate implements the Provider interface - OIDC doesn't use this directly
func (p *OIDCProviderImpl) Authenticate(ctx context.Context, credentials map[string]interface{}) (*auth.Claims, error) {
	// OIDC authentication is handled via GetAuthorizationURL + ExchangeCode + GetUserInfo
	// This method is not used for OIDC flow
	return nil, fmt.Errorf("use GetAuthorizationURL -> ExchangeCode -> GetUserInfo for OIDC flow")
}

// ProviderType returns the provider type
func (p *OIDCProviderImpl) ProviderType() auth.ProviderType {
	return p.providerType
}

// Name returns the provider name
func (p *OIDCProviderImpl) Name() string {
	return p.name
}

// Config returns the provider configuration
func (p *OIDCProviderImpl) Config() map[string]interface{} {
	return p.config
}

// GetAuthorizationURL builds the authorization URL for redirecting users to the OIDC provider
func (p *OIDCProviderImpl) GetAuthorizationURL(ctx context.Context, state string) (string, error) {
	if err := p.ensureDiscoveryDocument(ctx); err != nil {
		return "", err
	}

	clientID, ok := p.config["client_id"].(string)
	if !ok || clientID == "" {
		return "", auth.ErrProviderConfigInvalid
	}

	redirectURI, ok := p.config["redirect_uri"].(string)
	if !ok || redirectURI == "" {
		return "", auth.ErrProviderConfigInvalid
	}

	q := url.Values{
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
		"response_type": {"code"},
		"scope":         {p.getScopes()},
		"state":         {state},
	}

	return p.discoveryDoc.AuthorizationEndpoint + "?" + q.Encode(), nil
}

// ExchangeCode exchanges an authorization code for tokens
func (p *OIDCProviderImpl) ExchangeCode(ctx context.Context, code string) (*auth.OIDCTokenResponse, error) {
	if err := p.ensureDiscoveryDocument(ctx); err != nil {
		return nil, err
	}

	clientID, ok := p.config["client_id"].(string)
	if !ok || clientID == "" {
		return nil, auth.ErrProviderConfigInvalid
	}

	clientSecret, ok := p.config["client_secret"].(string)
	if !ok || clientSecret == "" {
		return nil, auth.ErrProviderConfigInvalid
	}

	redirectURI, ok := p.config["redirect_uri"].(string)
	if !ok || redirectURI == "" {
		return nil, auth.ErrProviderConfigInvalid
	}

	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.discoveryDoc.TokenEndpoint, nil)
	if err != nil {
		return nil, auth.ErrCodeExchangeFailed
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Body = io.NopCloser(strings.NewReader(data.Encode()))

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, auth.ErrCodeExchangeFailed
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, auth.ErrCodeExchangeFailed
	}

	var tokenResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, auth.ErrCodeExchangeFailed
	}

	return &auth.OIDCTokenResponse{
		AccessToken:  getStringField(tokenResp, "access_token"),
		TokenType:    getStringField(tokenResp, "token_type"),
		ExpiresIn:    getInt64Field(tokenResp, "expires_in"),
		RefreshToken: getStringField(tokenResp, "refresh_token"),
		IDToken:      getStringField(tokenResp, "id_token"),
	}, nil
}

// GetUserInfo fetches user information from the OIDC provider
func (p *OIDCProviderImpl) GetUserInfo(ctx context.Context, accessToken string) (*auth.OIDCUserInfo, error) {
	if err := p.ensureDiscoveryDocument(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.discoveryDoc.UserinfoEndpoint, nil)
	if err != nil {
		return nil, auth.ErrUserInfoFetchFailed
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, auth.ErrUserInfoFetchFailed
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, auth.ErrUserInfoFetchFailed
	}

	var userInfo auth.OIDCUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, auth.ErrUserInfoFetchFailed
	}

	return &userInfo, nil
}

// ensureDiscoveryDocument fetches the discovery document if not already loaded
func (p *OIDCProviderImpl) ensureDiscoveryDocument(ctx context.Context) error {
	if p.discoveryDoc != nil {
		return nil
	}

	discoveryURL, ok := p.config["discovery_url"].(string)
	if !ok || discoveryURL == "" {
		return auth.ErrProviderConfigInvalid
	}

	doc, err := p.discoveryService.GetDiscoveryDocument(ctx, discoveryURL)
	if err != nil {
		return err
	}

	p.discoveryDoc = doc
	return nil
}

// getScopes returns the OIDC scopes for this provider
func (p *OIDCProviderImpl) getScopes() string {
	if scopes, ok := p.config["scopes"].(string); ok && scopes != "" {
		return scopes
	}
	return "openid profile email"
}

// Helper functions
func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt64Field(m map[string]interface{}, key string) int64 {
	switch v := m[key].(type) {
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case int64:
		return v
	}
	return 0
}
