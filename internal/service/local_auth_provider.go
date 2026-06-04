package service

import (
	"context"

	"github.com/shashisharma307703/vedantam/internal/domain/auth"
	"golang.org/x/crypto/bcrypt"
)

// LocalAuthProviderImpl implements the LocalProvider interface for local username/password auth
type LocalAuthProviderImpl struct {
	name         string
	providerType auth.ProviderType
	config       map[string]interface{}
	bcryptCost   int
}

// NewLocalAuthProvider creates a new local auth provider instance
func NewLocalAuthProvider(config map[string]interface{}, bcryptCost int) *LocalAuthProviderImpl {
	return &LocalAuthProviderImpl{
		name:         "Local Auth",
		providerType: auth.ProviderTypeLocal,
		config:       config,
		bcryptCost:   bcryptCost,
	}
}

// Authenticate authenticates a user with username/password
func (p *LocalAuthProviderImpl) Authenticate(ctx context.Context, credentials map[string]interface{}) (*auth.Claims, error) {
	// Credentials should contain: username/email and password
	// This is typically used during login, but actual user lookup and verification
	// happens in the AuthService
	return nil, auth.ErrInvalidCredentials
}

// ProviderType returns the provider type
func (p *LocalAuthProviderImpl) ProviderType() auth.ProviderType {
	return p.providerType
}

// Name returns the provider name
func (p *LocalAuthProviderImpl) Name() string {
	return p.name
}

// Config returns the provider configuration
func (p *LocalAuthProviderImpl) Config() map[string]interface{} {
	return p.config
}

// VerifyPassword verifies a password against a bcrypt hash
func (p *LocalAuthProviderImpl) VerifyPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return auth.ErrInvalidCredentials
	}
	return nil
}

// HashPassword hashes a password using bcrypt
func (p *LocalAuthProviderImpl) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), p.bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
