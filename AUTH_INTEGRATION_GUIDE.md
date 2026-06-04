# Vedsutra Authentication System - Integration Guide

## Quick Start

### 1. Run Database Migrations
```bash
# Execute the migration to create auth tables
psql -U your_db_user -d school_erp -f db/migrations/001_auth_schema.sql
```

### 2. Configure Environment
Create `.env` file or set environment variables:
```bash
# Required for OIDC
OIDC_DISCOVERY_URL=https://your-keycloak.com/realms/vedsutra/.well-known/openid-configuration
OIDC_CLIENT_ID=vedsutra-backend
OIDC_CLIENT_SECRET=your-secret

# JWT
JWT_SECRET=your-super-secret-key  # Will be auto-generated if not provided
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=168h

# Local Auth (optional)
LOCAL_AUTH_ENABLED=true
PASSWORD_MIN_LENGTH=8
BCRYPT_COST=10
```

### 3. Start the Application
```bash
go run cmd/api/main.go
```

## Architecture Components

### Multi-Tenant Support
- **Tenant Resolution:** Via subdomain (e.g., `corp25.myapp.com`)
- **Middleware:** `TenantMiddleware` extracts tenant early in request lifecycle
- **Isolation:** All queries include `tenant_id` for data isolation

### JWT Token Structure
```json
{
  "user_id": 123,
  "tenant_id": 1,
  "tenant_slug": "corp25",
  "email": "user@example.com",
  "username": "user",
  "roles": ["admin", "teacher"],
  "permissions": ["read:classes", "write:grades"],
  "modules": ["attendance", "grades", "fees"],
  "token_type": "access",
  "iat": 1234567890,
  "exp": 1234568790,
  "iss": "vedsutra"
}
```

### Auth Flow Diagrams

#### Local Authentication
```
User → POST /auth/login
    ├→ Validate credentials against users table
    ├→ Hash password with bcrypt
    └→ Generate JWT tokens & refresh session
    ← Access Token + Refresh Token
```

#### OIDC (Keycloak) Authentication
```
User → GET /auth/sso
    ├→ Generate secure state parameter
    ├→ Fetch OIDC discovery document
    ├→ Build authorization URL
    └→ Redirect to Keycloak

User Logs in at Keycloak
    ↓
Keycloak → GET /sso/callback?code=...&state=...
    ├→ Validate state parameter
    ├→ Exchange code for tokens
    ├→ Fetch user info from Keycloak
    ├→ JIT provision user in local DB
    ├→ Generate local JWT tokens
    └→ Create refresh session
    ← Redirect to frontend with tokens
```

#### Token Refresh
```
Frontend → POST /auth/refresh (with refresh_token)
    ├→ Validate refresh token
    ├→ Check if revoked/expired
    ├→ If refresh_token_rotation=true:
    │   ├→ Revoke old session
    │   ├→ Generate new refresh token
    │   └→ Create new session
    └→ Generate new access token
    ← New Access Token (+ optionally new Refresh Token)
```

## Database Schema

### auth_sessions Table
Stores refresh token sessions with security metadata:
```sql
CREATE TABLE auth_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL,
    refresh_token_hash VARCHAR(255) NOT NULL UNIQUE,  -- SHA256 hash
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP,
    ip_address VARCHAR(45),
    user_agent TEXT,
    is_revoked BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);
```

### auth_providers Table
Per-tenant authentication provider configuration:
```sql
CREATE TABLE auth_providers (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    provider_type VARCHAR(50) NOT NULL,
    config JSONB,  -- { discovery_url, client_id, client_secret, scopes, redirect_uri }
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    UNIQUE(tenant_id, provider_type)
);
```

### oidc_discovery_cache Table
Caches OIDC discovery documents to avoid repeated fetching:
```sql
CREATE TABLE oidc_discovery_cache (
    id BIGSERIAL PRIMARY KEY,
    issuer VARCHAR(255) NOT NULL UNIQUE,
    discovery_doc JSONB,  -- Full OIDC discovery document
    cached_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);
```

## Important: Fix Required Before Production

### SQL.DB Adapter Issue
The repositories are designed for `database/sql.DB`, but the application uses `pgxpool.Pool`:

**Current (won't work):**
```go
sqlDB := &sql.DB{}  // Empty placeholder
authSessionRepo := repository.NewAuthSessionRepository(sqlDB)  // Won't work!
```

**Solution - Choose One:**

#### Option A: Refactor to pgxpool (Recommended)
```go
// Change repositories to use pgxpool directly
type authSessionRepo struct {
    pool *pgxpool.Pool  // Instead of sql.DB
}

// Implement methods using pool.QueryRowContext, pool.Exec, etc.
```

#### Option B: SQL Adapter
```go
// Use stdlib sql package wrapper
import "database/sql"
import "github.com/jackc/pgx/v5/stdlib"

// Create SQL connection from pgxpool
connector := stdlib.GetDefaultDriver().OpenConnector("dbname=school_erp user=sms_admin")
sqlDB := sql.OpenDB(connector)

// Now pass to repositories
authSessionRepo := repository.NewAuthSessionRepository(sqlDB)
```

**Status:** Code compiles but application will panic at runtime until this is fixed.

## Implementation Status

### ✅ Completed
- [x] Domain models and interfaces
- [x] Configuration layer
- [x] Repository interfaces and implementations
- [x] Token service (JWT generation/validation)
- [x] OIDC discovery service
- [x] Local and OIDC auth providers
- [x] Auth service orchestration
- [x] HTTP handlers for all endpoints
- [x] Middleware (tenant, auth, CORS)
- [x] Database schema (migration file)
- [x] Dependency injection in app bootstrap

### ⚠️ TODO Before Production
- [ ] Fix SQL.DB/pgxpool adapter
- [ ] Load roles/permissions from database
- [ ] Implement Keycloak testing
- [ ] Add comprehensive unit tests
- [ ] Add integration tests
- [ ] Implement rate limiting
- [ ] Add request validation
- [ ] Audit logging
- [ ] Redis cache for OIDC state (instead of in-memory)
- [ ] Implement logout with cookie clearing
- [ ] Add password reset flow
- [ ] Add MFA support
- [ ] Add account lockout logic

## Testing the System

### Test Local Login
```bash
# 1. Create a test user in database
INSERT INTO users (tenant_id, email, username, password, is_active)
VALUES (1, 'test@example.com', 'test@example.com', '$2a$10$...bcrypt_hash...', true);

# 2. Login
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Slug: default" \
  -d '{"username":"test@example.com","password":"testpass123"}'

# Should return access_token and refresh_token
```

### Test Protected Endpoint
```bash
curl -X GET http://localhost:8080/auth/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "X-Tenant-Slug: default"

# Should return user info
```

### Test Token Refresh
```bash
curl -X POST http://localhost:8080/auth/refresh \
  -H "Authorization: Bearer YOUR_REFRESH_TOKEN" \
  -H "X-Tenant-Slug: default"

# Should return new access token
```

## Best Practices

1. **Environment Variables:** Never commit secrets to version control
2. **HTTPS:** Use TLS in production (secure cookies, secure headers)
3. **Token Storage:** 
   - Frontend: Store tokens in httpOnly cookies or secure storage
   - Never expose refresh tokens in URL parameters
4. **Rate Limiting:** Implement rate limits on `/auth/login` endpoint
5. **Audit Logging:** Log all auth events (login, logout, token refresh, failures)
6. **Monitoring:** Monitor failed login attempts and anomalies
7. **State Management:** Use Redis for distributed systems (not in-memory)
8. **Certificate Pinning:** For mobile apps using OIDC

## Support & Maintenance

### Regular Tasks
- Monitor failed login attempts
- Review audit logs
- Update dependencies (esp. JWT library)
- Rotate secrets periodically
- Clean up expired sessions

### Troubleshooting

**"Token validation failed"**
- Check JWT_SECRET matches between token generation and validation
- Verify token hasn't expired
- Check token format (Bearer <token>)

**"OIDC discovery failed"**
- Verify OIDC_DISCOVERY_URL is correct
- Check network connectivity to Keycloak
- Verify discovery document is accessible

**"Tenant not found"**
- Check subdomain matches tenant in database
- Verify X-Tenant-Slug header if not using subdomains
- Ensure tenant record exists

---

**System Status:** ✅ Ready for Integration & Testing
**Next Action:** Fix SQL.DB adapter and run against real database
