package auth

import "context"

// Provider defines the interface for authentication providers
// This enables the Strategy Pattern for pluggable auth implementations
type Provider interface {
	// Authenticate authenticates a user and returns claims
	Authenticate(ctx context.Context, credentials map[string]interface{}) (*Claims, error)

	// ProviderType returns the provider type
	ProviderType() ProviderType

	// Name returns a human-readable name
	Name() string

	// Config returns the provider's current configuration
	Config() map[string]interface{}
}

// OIDCProvider defines additional methods for OIDC-based providers
type OIDCProvider interface {
	Provider

	// GetAuthorizationURL builds the authorization URL for redirecting users
	GetAuthorizationURL(ctx context.Context, state string) (string, error)

	// ExchangeCode exchanges an authorization code for tokens
	ExchangeCode(ctx context.Context, code string) (*OIDCTokenResponse, error)

	// GetUserInfo fetches user information from the OIDC provider
	GetUserInfo(ctx context.Context, accessToken string) (*OIDCUserInfo, error)
}

// OIDCTokenResponse represents the response from an OIDC token endpoint
type OIDCTokenResponse struct {
	AccessToken  string
	TokenType    string
	ExpiresIn    int64
	RefreshToken string
	IDToken      string
}

// LocalProvider defines additional methods for local authentication
type LocalProvider interface {
	Provider

	// VerifyPassword verifies a password against a hash
	VerifyPassword(password, hash string) error

	// HashPassword hashes a password
	HashPassword(password string) (string, error)
}
