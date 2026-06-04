package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shashisharma307703/vedantam/config"
	"github.com/shashisharma307703/vedantam/internal/domain/auth"
)

// TokenService handles JWT token generation and validation
type TokenService interface {
	// GenerateAccessToken creates a new access token
	GenerateAccessToken(claims *auth.Claims) (*auth.Token, error)

	// GenerateRefreshToken creates a new refresh token
	GenerateRefreshToken(claims *auth.Claims) (*auth.Token, error)

	// ValidateToken validates a JWT token string
	ValidateToken(tokenString string) (*auth.Claims, error)

	// RefreshAccessToken generates a new access token from claims
	RefreshAccessToken(claims *auth.Claims) (*auth.Token, error)
}

type tokenService struct {
	config *config.AuthConfig
}

// NewTokenService creates a new token service
func NewTokenService(cfg *config.AuthConfig) TokenService {
	return &tokenService{
		config: cfg,
	}
}

// GenerateAccessToken creates a new access token
func (s *tokenService) GenerateAccessToken(claims *auth.Claims) (*auth.Token, error) {
	claims.TokenType = auth.TokenTypeAccess
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(s.config.JWTAccessTokenExpiry))
	claims.IssuedAt = jwt.NewNumericDate(time.Now())

	tokenString, err := s.signToken(claims)
	if err != nil {
		return nil, err
	}

	return &auth.Token{
		Type:      auth.TokenTypeAccess,
		Value:     tokenString,
		ExpiresAt: claims.ExpiresAt.Time,
		Claims:    *claims,
	}, nil
}

// GenerateRefreshToken creates a new refresh token
func (s *tokenService) GenerateRefreshToken(claims *auth.Claims) (*auth.Token, error) {
	claims.TokenType = auth.TokenTypeRefresh
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(s.config.JWTRefreshTokenExpiry))
	claims.IssuedAt = jwt.NewNumericDate(time.Now())

	tokenString, err := s.signToken(claims)
	if err != nil {
		return nil, err
	}

	return &auth.Token{
		Type:      auth.TokenTypeRefresh,
		Value:     tokenString,
		ExpiresAt: claims.ExpiresAt.Time,
		Claims:    *claims,
	}, nil
}

// ValidateToken validates a JWT token string
func (s *tokenService) ValidateToken(tokenString string) (*auth.Claims, error) {
	claims := &auth.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != s.config.JWTAlgorithm {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, auth.ErrInvalidToken
	}

	if !token.Valid {
		return nil, auth.ErrInvalidToken
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token from claims
func (s *tokenService) RefreshAccessToken(claims *auth.Claims) (*auth.Token, error) {
	newClaims := &auth.Claims{
		UserID:      claims.UserID,
		TenantID:    claims.TenantID,
		TenantSlug:  claims.TenantSlug,
		Email:       claims.Email,
		Username:    claims.Username,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
		Modules:     claims.Modules,
	}
	return s.GenerateAccessToken(newClaims)
}

// signToken signs a JWT token with the configured algorithm
func (s *tokenService) signToken(claims *auth.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", auth.ErrJWTSigningFailed
	}
	return tokenString, nil
}

// HashRefreshToken creates a SHA256 hash of a refresh token for secure storage
func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
