-- Migration: 001_auth_schema.sql
-- Description: Create authentication-related tables for OIDC and local auth

-- auth_sessions table: Store refresh token sessions
CREATE TABLE IF NOT EXISTS auth_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL,
    refresh_token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP,
    ip_address VARCHAR(45),
    user_agent TEXT,
    is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_auth_sessions_user_id ON auth_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_tenant_id ON auth_sessions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_refresh_token_hash ON auth_sessions(refresh_token_hash);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires_at ON auth_sessions(expires_at);

-- auth_providers table: Configure authentication providers per tenant
CREATE TABLE IF NOT EXISTS auth_providers (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    provider_type VARCHAR(50) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    UNIQUE(tenant_id, provider_type)
);

CREATE INDEX IF NOT EXISTS idx_auth_providers_tenant_id ON auth_providers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_auth_providers_provider_type ON auth_providers(provider_type);

-- oidc_discovery_cache table: Cache OIDC discovery documents
CREATE TABLE IF NOT EXISTS oidc_discovery_cache (
    id BIGSERIAL PRIMARY KEY,
    issuer VARCHAR(255) NOT NULL UNIQUE,
    discovery_doc JSONB NOT NULL,
    cached_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    FOREIGN KEY (expires_at) REFERENCES auth_providers(created_at) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_oidc_discovery_cache_issuer ON oidc_discovery_cache(issuer);
CREATE INDEX IF NOT EXISTS idx_oidc_discovery_cache_expires_at ON oidc_discovery_cache(expires_at);

-- Optional: Add OIDC provider columns to existing users table if needed
-- ALTER TABLE users ADD COLUMN IF NOT EXISTS oidc_subject VARCHAR(255);
-- ALTER TABLE users ADD COLUMN IF NOT EXISTS oidc_provider VARCHAR(50);
-- CREATE INDEX IF NOT EXISTS idx_users_oidc_subject ON users(oidc_subject);

-- Optional: Add password hash column to users if using local auth
-- ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);
