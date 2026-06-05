-- name: CreateAuthProvider :one
INSERT INTO auth_providers (tenant_id, provider_type, config, is_active)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAuthProvider :one
SELECT * FROM auth_providers WHERE id = $1;

-- name: GetAuthProviderByTenantAndType :one
SELECT * FROM auth_providers 
WHERE tenant_id = $1 AND provider_type = $2 AND is_active = true;

-- name: ListActiveAuthProvidersByTenant :many
SELECT * FROM auth_providers 
WHERE tenant_id = $1 AND is_active = true 
ORDER BY provider_type;

-- name: UpdateAuthProvider :exec
UPDATE auth_providers 
SET provider_type = $2, config = $3, is_active = $4, updated_at = NOW()
WHERE id = $1;

-- name: DeleteAuthProvider :exec
DELETE FROM auth_providers WHERE id = $1;