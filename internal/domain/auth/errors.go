package auth

import "errors"

// Auth-specific errors
var (
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrUserNotFound            = errors.New("user not found")
	ErrTokenExpired            = errors.New("token expired")
	ErrInvalidToken            = errors.New("invalid token")
	ErrInvalidState            = errors.New("invalid state parameter")
	ErrProviderNotFound        = errors.New("provider not found")
	ErrProviderNotSupported    = errors.New("provider not supported")
	ErrProviderConfigInvalid   = errors.New("invalid provider configuration")
	ErrRefreshTokenRevoked     = errors.New("refresh token revoked")
	ErrRefreshTokenExpired     = errors.New("refresh token expired")
	ErrJWTSigningFailed        = errors.New("JWT signing failed")
	ErrOIDCDiscoveryFailed     = errors.New("OIDC discovery failed")
	ErrCodeExchangeFailed      = errors.New("code exchange failed")
	ErrUserInfoFetchFailed     = errors.New("failed to fetch user info")
	ErrTenantNotFound          = errors.New("tenant not found")
	ErrTenantMismatch          = errors.New("tenant mismatch")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrForbidden               = errors.New("forbidden")
	ErrSessionNotFound         = errors.New("session not found")
)
