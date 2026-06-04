package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/shashisharma307703/vedantam/internal/domain/auth"
	"github.com/shashisharma307703/vedantam/internal/logger"
	"github.com/shashisharma307703/vedantam/internal/service"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService service.AuthService
	logger      logger.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService service.AuthService, logger logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// RegisterRoutes registers all auth routes
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/login", h.HandleLogin)
	r.Get("/auth/sso", h.HandleSSO)
	r.Get("/sso/callback", h.HandleSSOCallback)
	r.Post("/auth/refresh", h.HandleRefreshToken)
	r.Get("/auth/me", h.HandleUserInfo)
	r.Post("/auth/logout", h.HandleLogout)
}

// LoginRequest represents a login request payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// TokenResponse represents a token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"description,omitempty"`
}

// HandleLogin handles local username/password login
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "Failed to parse request")
		return
	}

	// Get tenant from context (set by middleware)
	tenant, ok := ctx.Value("tenant").(*auth.Tenant)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "missing_tenant", "Tenant not found")
		return
	}

	// Get client IP and user agent
	ipAddress := r.Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = strings.Split(r.RemoteAddr, ":")[0]
	}
	userAgent := r.Header.Get("User-Agent")

	// Authenticate
	loginReq := &auth.LoginRequest{
		Username: req.Username,
		Password: req.Password,
		TenantID: tenant.ID,
	}
	resp, err := h.authService.LoginWithCredentials(ctx, loginReq, ipAddress, userAgent)
	if err != nil {
		h.logger.Warnf("login failed: %v", err)
		h.respondError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid username or password")
		return
	}

	h.respondSuccess(w, http.StatusOK, &TokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
		TokenType:    resp.TokenType,
	})
}

// HandleSSO initiates OIDC/SSO login flow
func (h *AuthHandler) HandleSSO(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get tenant from context
	tenant, ok := ctx.Value("tenant").(*auth.Tenant)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "missing_tenant", "Tenant not found")
		return
	}

	// Get OIDC authorization URL
	authURL, err := h.authService.GetOIDCAuthorizationURL(ctx, tenant.ID)
	if err != nil {
		h.logger.Errorf("failed to get OIDC URL: %v", err)
		h.respondError(w, http.StatusInternalServerError, "server_error", "Failed to initiate SSO")
		return
	}

	// Redirect to OIDC provider
	http.Redirect(w, r, authURL, http.StatusFound)
}

// HandleSSOCallback handles OIDC provider callback
func (h *AuthHandler) HandleSSOCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse callback parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorCode := r.URL.Query().Get("error")

	if errorCode != "" {
		h.respondError(w, http.StatusBadRequest, errorCode, r.URL.Query().Get("error_description"))
		return
	}

	if code == "" || state == "" {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "Missing code or state parameter")
		return
	}

	// Get tenant from context
	tenant, ok := ctx.Value("tenant").(*auth.Tenant)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "missing_tenant", "Tenant not found")
		return
	}

	// Get client IP and user agent
	ipAddress := r.Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = strings.Split(r.RemoteAddr, ":")[0]
	}
	userAgent := r.Header.Get("User-Agent")

	// Exchange code for tokens
	resp, err := h.authService.ExchangeOIDCCode(ctx, code, state, tenant.Slug, ipAddress, userAgent)
	if err != nil {
		h.logger.Warnf("OIDC code exchange failed: %v", err)
		h.respondError(w, http.StatusUnauthorized, "invalid_code", "Failed to exchange authorization code")
		return
	}

	// Redirect to frontend with tokens (in production, use secure methods)
	redirectURL := "http://localhost:3000/auth/callback?access_token=" + resp.AccessToken +
		"&refresh_token=" + resp.RefreshToken + "&token_type=" + resp.TokenType

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// HandleRefreshToken handles token refresh
func (h *AuthHandler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get refresh token from request or cookie
	var refreshToken string

	// Try to get from Authorization header first
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		refreshToken = authHeader[7:]
	} else {
		// Try to get from request body
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			refreshToken = req.RefreshToken
		}
	}

	if refreshToken == "" {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "Missing refresh token")
		return
	}

	// Refresh token
	resp, err := h.authService.RefreshAccessToken(ctx, refreshToken)
	if err != nil {
		h.logger.Warnf("token refresh failed: %v", err)
		h.respondError(w, http.StatusUnauthorized, "invalid_token", "Failed to refresh token")
		return
	}

	h.respondSuccess(w, http.StatusOK, &TokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
		TokenType:    resp.TokenType,
	})
}

// HandleUserInfo returns current user information
func (h *AuthHandler) HandleUserInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get auth context from middleware
	authCtx, ok := ctx.Value("auth").(*auth.AuthContext)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	h.respondSuccess(w, http.StatusOK, map[string]interface{}{
		"user_id":     authCtx.UserID,
		"tenant_id":   authCtx.TenantID,
		"tenant_slug": authCtx.TenantSlug,
		"email":       authCtx.Email,
		"roles":       authCtx.Roles,
		"permissions": authCtx.Permissions,
		"modules":     authCtx.Modules,
	})
}

// HandleLogout handles user logout
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get auth context
	authCtx, ok := ctx.Value("auth").(*auth.AuthContext)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	// Revoke all user sessions
	if err := h.authService.RevokeAllUserSessions(ctx, authCtx.UserID); err != nil {
		h.logger.Errorf("failed to revoke sessions: %v", err)
		h.respondError(w, http.StatusInternalServerError, "server_error", "Failed to logout")
		return
	}

	h.respondSuccess(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

// Helper methods for responses

func (h *AuthHandler) respondSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *AuthHandler) respondError(w http.ResponseWriter, statusCode int, errorCode, description string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(&ErrorResponse{
		Error:       errorCode,
		Description: description,
	})
}
