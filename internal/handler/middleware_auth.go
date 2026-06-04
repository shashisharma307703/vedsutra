package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/shashisharma307703/vedantam/internal/domain/auth"
	"github.com/shashisharma307703/vedantam/internal/logger"
	"github.com/shashisharma307703/vedantam/internal/service"
)

// TenantMiddleware extracts tenant information from the request (subdomain or header)
// This should be applied early in the middleware chain
func TenantMiddleware(logger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract tenant from Host header
			host := r.Host
			parts := strings.Split(host, ".")

			var tenantSlug string
			if len(parts) > 1 {
				// Multi-tenant via subdomain: tenant.example.com
				tenantSlug = parts[0]
			} else {
				// Single tenant or use header
				tenantSlug = r.Header.Get("X-Tenant-Slug")
				if tenantSlug == "" {
					// Default to localhost for development
					tenantSlug = "default"
				}
			}

			// Create tenant object (in production, fetch from database)
			tenant := &auth.Tenant{
				ID:   1, // TODO: Fetch from database based on slug
				Slug: tenantSlug,
				Name: tenantSlug,
			}

			// Add tenant to context
			ctx := context.WithValue(r.Context(), "tenant", tenant)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthMiddleware validates JWT token from Authorization header
// This should be applied to protected routes
func AuthMiddleware(authService service.AuthService, logger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondError(w, http.StatusUnauthorized, "missing_token", "Authorization header is required")
				return
			}

			// Bearer token format
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondError(w, http.StatusUnauthorized, "invalid_token_format", "Expected 'Bearer <token>'")
				return
			}

			token := parts[1]

			// Validate token
			authCtx, err := authService.ValidateToken(r.Context(), token)
			if err != nil {
				logger.Warnf("token validation failed: %v", err)
				respondError(w, http.StatusUnauthorized, "invalid_token", "Token validation failed")
				return
			}

			// Add auth context to request
			ctx := context.WithValue(r.Context(), "auth", authCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthMiddleware validates JWT token if present, but doesn't fail if missing
// Useful for endpoints that should work with or without authentication
func OptionalAuthMiddleware(authService service.AuthService, logger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to extract token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// No token, continue
				next.ServeHTTP(w, r)
				return
			}

			// Bearer token format
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				// Invalid format, skip auth
				next.ServeHTTP(w, r)
				return
			}

			token := parts[1]

			// Try to validate token
			authCtx, err := authService.ValidateToken(r.Context(), token)
			if err != nil {
				logger.Debugf("optional auth: token validation failed: %v", err)
				// Continue without auth context
				next.ServeHTTP(w, r)
				return
			}

			// Add auth context to request
			ctx := context.WithValue(r.Context(), "auth", authCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CORSMiddleware handles CORS headers
func CORSMiddleware(allowedOrigins string) func(http.Handler) http.Handler {
	origins := strings.Split(allowedOrigins, ",")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			for _, allowed := range origins {
				if strings.TrimSpace(allowed) == origin || strings.TrimSpace(allowed) == "*" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-Slug")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Access-Control-Max-Age", "3600")
					break
				}
			}

			// Handle preflight
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper function for error responses
func respondError(w http.ResponseWriter, statusCode int, errorCode, description string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp := map[string]interface{}{
		"error": errorCode,
	}
	if description != "" {
		resp["description"] = description
	}
	json.NewEncoder(w).Encode(resp)
}
