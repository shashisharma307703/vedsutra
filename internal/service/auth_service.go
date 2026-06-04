package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/shashisharma307703/vedantam/config"
	"github.com/shashisharma307703/vedantam/internal/domain/auth"
	"github.com/shashisharma307703/vedantam/internal/repository"
)

// AuthService is the main authentication service orchestrating all auth flows
type AuthService interface {
	// Local Auth
	LoginWithCredentials(ctx context.Context, req *auth.LoginRequest, ipAddress, userAgent string) (*auth.LoginResponse, error)

	// OIDC
	GetOIDCAuthorizationURL(ctx context.Context, tenantID int64) (string, error)
	ExchangeOIDCCode(ctx context.Context, code, state, tenantID string, ipAddress, userAgent string) (*auth.LoginResponse, error)

	// Token Management
	RefreshAccessToken(ctx context.Context, refreshToken string) (*auth.LoginResponse, error)
	ValidateToken(ctx context.Context, token string) (*auth.AuthContext, error)
	RevokeToken(ctx context.Context, sessionID int64) error
	RevokeAllUserSessions(ctx context.Context, userID int64) error

	// State Management
	GenerateOIDCState(tenantID int64) (string, error)
	ValidateOIDCState(state string) (int64, error)
}

type authService struct {
	config                 *config.AuthConfig
	tokenService           TokenService
	oidcDiscoveryService   OIDCDiscoveryService
	authSessionRepo        repository.AuthSessionRepository
	authProviderRepo       repository.AuthProviderRepository
	userRepo               repository.UserRepository
	oidcProviders          map[auth.ProviderType]auth.OIDCProvider
	localProvider          auth.LocalProvider
	stateStore             map[string]stateEntry // In-memory state store (consider using Redis for production)
}

type stateEntry struct {
	TenantID  int64
	ExpiresAt time.Time
}

// NewAuthService creates a new authentication service
func NewAuthService(
	cfg *config.AuthConfig,
	tokenService TokenService,
	oidcDiscoveryService OIDCDiscoveryService,
	authSessionRepo repository.AuthSessionRepository,
	authProviderRepo repository.AuthProviderRepository,
	userRepo repository.UserRepository,
) AuthService {
	return &authService{
		config:               cfg,
		tokenService:         tokenService,
		oidcDiscoveryService: oidcDiscoveryService,
		authSessionRepo:      authSessionRepo,
		authProviderRepo:     authProviderRepo,
		userRepo:             userRepo,
		oidcProviders:        make(map[auth.ProviderType]auth.OIDCProvider),
		localProvider:        NewLocalAuthProvider(nil, cfg.BcryptCost),
		stateStore:           make(map[string]stateEntry),
	}
}

// LoginWithCredentials authenticates a user with username/password
func (s *authService) LoginWithCredentials(ctx context.Context, req *auth.LoginRequest, ipAddress, userAgent string) (*auth.LoginResponse, error) {
	// Get user by username or email
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, auth.ErrInvalidCredentials
	}

	// Verify password
	if err := s.localProvider.VerifyPassword(req.Password, user.Password); err != nil {
		return nil, err
	}

	// Generate tokens
	return s.generateLoginResponse(ctx, user, ipAddress, userAgent)
}

// GetOIDCAuthorizationURL builds the OIDC authorization URL for a tenant
func (s *authService) GetOIDCAuthorizationURL(ctx context.Context, tenantID int64) (string, error) {
	// Get OIDC provider config for tenant
	provider, err := s.authProviderRepo.GetByTenantAndType(ctx, tenantID, auth.ProviderTypeKeycloak)
	if err != nil {
		return "", auth.ErrProviderNotFound
	}

	// Get or create OIDC provider instance
	oidcProvider, err := s.getOrCreateOIDCProvider(ctx, provider)
	if err != nil {
		return "", err
	}

	// Generate state
	state, err := s.GenerateOIDCState(tenantID)
	if err != nil {
		return "", err
	}

	// Get authorization URL
	return oidcProvider.GetAuthorizationURL(ctx, state)
}

// ExchangeOIDCCode exchanges an authorization code for tokens (OIDC callback)
func (s *authService) ExchangeOIDCCode(ctx context.Context, code, state, tenantID string, ipAddress, userAgent string) (*auth.LoginResponse, error) {
	// Validate state
	tid, err := s.ValidateOIDCState(state)
	if err != nil {
		return nil, auth.ErrInvalidState
	}

	// Verify tenant ID matches
	// (In production, parse tenantID as int64 and compare)

	// Get OIDC provider config
	provider, err := s.authProviderRepo.GetByTenantAndType(ctx, tid, auth.ProviderTypeKeycloak)
	if err != nil {
		return nil, auth.ErrProviderNotFound
	}

	// Get OIDC provider instance
	oidcProvider, err := s.getOrCreateOIDCProvider(ctx, provider)
	if err != nil {
		return nil, err
	}

	// Exchange code for tokens
	tokenResp, err := oidcProvider.ExchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// Get user info
	userInfo, err := oidcProvider.GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}

	// JIT user provisioning: create or update user
	user, err := s.provisionUser(ctx, userInfo, tid)
	if err != nil {
		return nil, err
	}

	// Generate login response
	return s.generateLoginResponse(ctx, user, ipAddress, userAgent)
}

// RefreshAccessToken generates a new access token from a refresh token
func (s *authService) RefreshAccessToken(ctx context.Context, refreshToken string) (*auth.LoginResponse, error) {
	// Validate refresh token
	claims, err := s.tokenService.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Get session
	tokenHash := HashRefreshToken(refreshToken)
	session, err := s.authSessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, auth.ErrRefreshTokenExpired
	}

	// Check if revoked or expired
	if session.IsRevoked || time.Now().After(session.ExpiresAt) {
		return nil, auth.ErrRefreshTokenRevoked
	}

	// Check if token has been rotated (if rotation is enabled)
	if s.config.RefreshTokenRotation {
		// Revoke old token
		session.IsRevoked = true
		_ = s.authSessionRepo.Update(ctx, session)

		// Generate new refresh token
		newRefreshToken, err := s.tokenService.GenerateRefreshToken(claims)
		if err != nil {
			return nil, err
		}

		// Save new session
		newSession := &auth.RefreshSession{
			UserID:           claims.UserID,
			TenantID:         claims.TenantID,
			RefreshTokenHash: HashRefreshToken(newRefreshToken.Value),
			ExpiresAt:        newRefreshToken.ExpiresAt,
		}
		newSession, err = s.authSessionRepo.Create(ctx, newSession)
		if err != nil {
			return nil, err
		}

		// Generate new access token
		accessToken, err := s.tokenService.GenerateAccessToken(claims)
		if err != nil {
			return nil, err
		}

		return &auth.LoginResponse{
			AccessToken:  accessToken.Value,
			RefreshToken: newRefreshToken.Value,
			ExpiresIn:    int64(s.config.JWTAccessTokenExpiry.Seconds()),
			TokenType:    "Bearer",
		}, nil
	}

	// Non-rotated flow: just generate new access token
	newAccessToken, err := s.tokenService.GenerateAccessToken(claims)
	if err != nil {
		return nil, err
	}

	return &auth.LoginResponse{
		AccessToken:  newAccessToken.Value,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.config.JWTAccessTokenExpiry.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// ValidateToken validates a JWT token and returns auth context
func (s *authService) ValidateToken(ctx context.Context, token string) (*auth.AuthContext, error) {
	claims, err := s.tokenService.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	return &auth.AuthContext{
		UserID:      claims.UserID,
		TenantID:    claims.TenantID,
		TenantSlug:  claims.TenantSlug,
		Email:       claims.Email,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
		Modules:     claims.Modules,
	}, nil
}

// RevokeToken revokes a specific refresh token session
func (s *authService) RevokeToken(ctx context.Context, sessionID int64) error {
	return s.authSessionRepo.Revoke(ctx, sessionID)
}

// RevokeAllUserSessions revokes all sessions for a user
func (s *authService) RevokeAllUserSessions(ctx context.Context, userID int64) error {
	return s.authSessionRepo.RevokeAllForUser(ctx, userID)
}

// GenerateOIDCState generates a secure state parameter for OIDC flow
func (s *authService) GenerateOIDCState(tenantID int64) (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)
	s.stateStore[state] = stateEntry{
		TenantID:  tenantID,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	return state, nil
}

// ValidateOIDCState validates an OIDC state parameter
func (s *authService) ValidateOIDCState(state string) (int64, error) {
	entry, exists := s.stateStore[state]
	if !exists {
		return 0, auth.ErrInvalidState
	}

	if time.Now().After(entry.ExpiresAt) {
		delete(s.stateStore, state)
		return 0, auth.ErrInvalidState
	}

	delete(s.stateStore, state)
	return entry.TenantID, nil
}

// Private helper methods

func (s *authService) generateLoginResponse(ctx context.Context, user *auth.User, ipAddress, userAgent string) (*auth.LoginResponse, error) {
	// Create JWT claims
	claims := &auth.Claims{
		UserID:      user.ID,
		TenantID:    user.TenantID,
		Email:       user.Email,
		Username:    user.Username,
		Roles:       []string{}, // TODO: Fetch from database
		Permissions: []string{}, // TODO: Fetch from database
		Modules:     []string{}, // TODO: Fetch from database
	}

	// Generate access token
	accessToken, err := s.tokenService.GenerateAccessToken(claims)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := s.tokenService.GenerateRefreshToken(claims)
	if err != nil {
		return nil, err
	}

	// Save refresh session
	session := &auth.RefreshSession{
		UserID:           user.ID,
		TenantID:         user.TenantID,
		RefreshTokenHash: HashRefreshToken(refreshToken.Value),
		ExpiresAt:        refreshToken.ExpiresAt,
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
	}
	_, err = s.authSessionRepo.Create(ctx, session)
	if err != nil {
		return nil, err
	}

	return &auth.LoginResponse{
		AccessToken:  accessToken.Value,
		RefreshToken: refreshToken.Value,
		ExpiresIn:    int64(s.config.JWTAccessTokenExpiry.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func (s *authService) getOrCreateOIDCProvider(ctx context.Context, provider *auth.AuthProvider) (auth.OIDCProvider, error) {
	// Check if already cached
	if cached, exists := s.oidcProviders[provider.ProviderType]; exists {
		return cached, nil
	}

	// Create new provider
	oidcProvider := NewOIDCProvider(
		provider.ProviderType.String(),
		provider.ProviderType,
		provider.Config,
		s.oidcDiscoveryService,
	)

	// Cache it
	s.oidcProviders[provider.ProviderType] = oidcProvider

	return oidcProvider, nil
}

func (s *authService) provisionUser(ctx context.Context, userInfo *auth.OIDCUserInfo, tenantID int64) (*auth.User, error) {
	// Try to find existing user by email
	user, err := s.userRepo.GetByEmailAndTenant(ctx, userInfo.Email, tenantID)
	if err == nil {
		// User exists, update if needed
		return user, nil
	}

	// User doesn't exist, create new user (JIT provisioning)
	user = &auth.User{
		TenantID:  tenantID,
		Email:     userInfo.Email,
		Username:  userInfo.Email, // Use email as username for OIDC users
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.userRepo.Create(ctx, user)
}
