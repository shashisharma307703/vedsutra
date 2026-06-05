package repository

import (
	"context"
	"encoding/json"

	"github.com/shashisharma307703/vedantam/db/dbgen"
	"github.com/shashisharma307703/vedantam/internal/domain/auth"
)

type authProviderRepo struct {
	*Repository
}

func NewAuthProviderRepository(repo *Repository) AuthProviderRepository {
	return &authProviderRepo{
		Repository: repo,
	}
}

func (r *authProviderRepo) Create(
	ctx context.Context,
	provider *auth.AuthProvider,
) (*auth.AuthProvider, error) {

	configJSON, err := json.Marshal(provider.Config)
	if err != nil {
		return nil, err
	}

	row, err := r.Queries.CreateAuthProvider(
		ctx,
		dbgen.CreateAuthProviderParams{
			TenantID:     provider.TenantID,
			ProviderType: string(provider.ProviderType),
			Config:       configJSON,
			IsActive:     provider.IsActive,
		},
	)
	if err != nil {
		return nil, err
	}

	provider.ID = row.ID
	provider.CreatedAt = row.CreatedAt.Time
	provider.UpdatedAt = row.UpdatedAt.Time

	return provider, nil
}

func (r *authProviderRepo) GetByID(ctx context.Context, id int64) (*auth.AuthProvider, error) {
	query := `
		SELECT id, tenant_id, provider_type, config, is_active, created_at, updated_at
		FROM auth_providers
		WHERE id = $1
	`
	provider := &auth.AuthProvider{}
	var configJSON []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&provider.ID, &provider.TenantID, &provider.ProviderType, &configJSON,
		&provider.IsActive, &provider.CreatedAt, &provider.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(configJSON, &provider.Config); err != nil {
		return nil, err
	}
	return provider, nil
}

func (r *authProviderRepo) GetByTenantAndType(ctx context.Context, tenantID int64, providerType auth.ProviderType) (*auth.AuthProvider, error) {
	query := `
		SELECT id, tenant_id, provider_type, config, is_active, created_at, updated_at
		FROM auth_providers
		WHERE tenant_id = $1 AND provider_type = $2 AND is_active = true
	`
	provider := &auth.AuthProvider{}
	var configJSON []byte
	err := r.db.QueryRowContext(ctx, query, tenantID, providerType).Scan(
		&provider.ID, &provider.TenantID, &provider.ProviderType, &configJSON,
		&provider.IsActive, &provider.CreatedAt, &provider.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(configJSON, &provider.Config); err != nil {
		return nil, err
	}
	return provider, nil
}

func (r *authProviderRepo) GetActiveByTenant(ctx context.Context, tenantID int64) ([]*auth.AuthProvider, error) {
	query := `
		SELECT id, tenant_id, provider_type, config, is_active, created_at, updated_at
		FROM auth_providers
		WHERE tenant_id = $1 AND is_active = true
		ORDER BY provider_type
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*auth.AuthProvider
	for rows.Next() {
		provider := &auth.AuthProvider{}
		var configJSON []byte
		err := rows.Scan(
			&provider.ID, &provider.TenantID, &provider.ProviderType, &configJSON,
			&provider.IsActive, &provider.CreatedAt, &provider.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(configJSON, &provider.Config); err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}
	return providers, rows.Err()
}

func (r *authProviderRepo) Update(ctx context.Context, provider *auth.AuthProvider) error {
	configJSON, err := json.Marshal(provider.Config)
	if err != nil {
		return err
	}

	query := `
		UPDATE auth_providers
		SET config = $1, is_active = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err = r.db.ExecContext(ctx, query, configJSON, provider.IsActive, provider.ID)
	return err
}

func (r *authProviderRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM auth_providers WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
