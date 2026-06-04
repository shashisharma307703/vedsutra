package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

// Claims defines the JWT claims for Vedsutra tokens
// Extends standard JWT claims with custom application claims
type Claims struct {
	UserID      int64    `json:"user_id"`
	TenantID    int64    `json:"tenant_id"`
	TenantSlug  string   `json:"tenant_slug"`
	Email       string   `json:"email"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	Modules     []string `json:"modules"`
	TokenType   TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// Valid validates the claims
func (c Claims) Valid() error {
	// RegisteredClaims will be validated by the JWT library
	// We just return nil here as the JWT parser will handle validation
	return nil
}
