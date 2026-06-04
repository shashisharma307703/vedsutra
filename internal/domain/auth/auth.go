package auth

import (
	"time"
)

// ProviderType defines supported authentication providers
type ProviderType string

const (
	ProviderTypeKeycloak ProviderType = "keycloak"
	ProviderTypeLocal    ProviderType = "local"
	ProviderTypeGoogle   ProviderType = "google"
	ProviderTypeMicrosoft ProviderType = "microsoft"
	ProviderTypeLDAP     ProviderType = "ldap"
)

// String() method for ProviderType
func (pt ProviderType) String() string {
	return string(pt)
}

// AuthProvider represents an authentication provider configuration
type AuthProvider struct {
	ID           int64
	TenantID     int64
	ProviderType ProviderType
	Config       map[string]interface{} // JSONB: oidc_endpoint, client_id, etc.
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// RefreshSession represents a refresh token session
type RefreshSession struct {
	ID               int64
	UserID           int64
	TenantID         int64
	RefreshTokenHash string       // SHA256 hash of the token
	ExpiresAt        time.Time
	CreatedAt        time.Time
	LastUsedAt       *time.Time
	IPAddress        string
	UserAgent        string
	IsRevoked        bool
}

// User represents a user in the system (basic auth info)
type User struct {
	ID        int64
	TenantID  int64
	Email     string
	Username  string
	Password  string // bcrypt hash
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// LoginRequest represents a login attempt
type LoginRequest struct {
	Username string
	Password string
	TenantID int64
}

// LoginResponse represents successful login response
type LoginResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	TokenType    string
}

// OIDCCallback represents data passed from OIDC provider callback
type OIDCCallback struct {
	Code      string
	State     string
	Error     string
	ErrorDesc string
}

// OIDCUserInfo represents user information from OIDC provider
type OIDCUserInfo struct {
	Subject   string                 `json:"sub"`
	Email     string                 `json:"email"`
	Name      string                 `json:"name"`
	Picture   string                 `json:"picture,omitempty"`
	Groups    []string               `json:"groups,omitempty"`
	Roles     []string               `json:"roles,omitempty"`
	Extra     map[string]interface{} `json:"-"`
}

// Token represents an authentication token
type Token struct {
	Type      TokenType
	Value     string
	ExpiresAt time.Time
	Claims    Claims
}

// TokenType differentiates between access and refresh tokens
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// Tenant represents tenant context
type Tenant struct {
	ID   int64
	Slug string
	Name string
}

// AuthContext contains authentication context for a request
type AuthContext struct {
	UserID      int64
	TenantID    int64
	TenantSlug  string
	Email       string
	Roles       []string
	Permissions []string
	Modules     []string
}
