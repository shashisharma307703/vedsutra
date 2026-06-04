package auth

import "time"

// OIDCDiscoveryDocument represents the OIDC provider's discovery document
// This is fetched from /.well-known/openid-configuration
type OIDCDiscoveryDocument struct {
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	UserinfoEndpoint      string   `json:"userinfo_endpoint"`
	JwksURI               string   `json:"jwks_uri"`
	ResponseTypesSupported []string `json:"response_types_supported"`
	ScopesSupported       []string `json:"scopes_supported"`
	GrantTypesSupported   []string `json:"grant_types_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported       []string `json:"claims_supported"`
	SubjectTypesSupported []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
}

// OIDCDiscoveryCache represents a cached discovery document
type OIDCDiscoveryCache struct {
	ID        int64
	Issuer    string
	Document  OIDCDiscoveryDocument
	CachedAt  time.Time
	ExpiresAt time.Time
}

// IsExpired checks if the cached discovery document has expired
func (c *OIDCDiscoveryCache) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// JWKSCache represents cached JWKS (JSON Web Key Set) from the OIDC provider
type JWKSCache struct {
	ID        int64
	Issuer    string
	Keys      []map[string]interface{} // JWKS keys
	CachedAt  time.Time
	ExpiresAt time.Time
}

// IsExpired checks if the cached JWKS has expired
func (j *JWKSCache) IsExpired() bool {
	return time.Now().After(j.ExpiresAt)
}
