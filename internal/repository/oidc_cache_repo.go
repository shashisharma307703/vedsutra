package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/shashisharma307703/vedantam/internal/domain/auth"
)

type oidcDiscoveryCacheRepo struct {
	db *sql.DB
}

func NewOIDCDiscoveryCacheRepository(db *sql.DB) OIDCDiscoveryCacheRepository {
	return &oidcDiscoveryCacheRepo{db: db}
}

func (r *oidcDiscoveryCacheRepo) Create(ctx context.Context, cache *auth.OIDCDiscoveryCache) (*auth.OIDCDiscoveryCache, error) {
	docJSON, err := json.Marshal(cache.Document)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO oidc_discovery_cache (issuer, discovery_doc, cached_at, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	err = r.db.QueryRowContext(ctx, query,
		cache.Issuer, docJSON, cache.CachedAt, cache.ExpiresAt,
	).Scan(&cache.ID)
	return cache, err
}

func (r *oidcDiscoveryCacheRepo) GetByIssuer(ctx context.Context, issuer string) (*auth.OIDCDiscoveryCache, error) {
	query := `
		SELECT id, issuer, discovery_doc, cached_at, expires_at
		FROM oidc_discovery_cache
		WHERE issuer = $1 AND expires_at > NOW()
	`
	cache := &auth.OIDCDiscoveryCache{}
	var docJSON []byte
	err := r.db.QueryRowContext(ctx, query, issuer).Scan(
		&cache.ID, &cache.Issuer, &docJSON, &cache.CachedAt, &cache.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(docJSON, &cache.Document); err != nil {
		return nil, err
	}
	return cache, nil
}

func (r *oidcDiscoveryCacheRepo) Update(ctx context.Context, cache *auth.OIDCDiscoveryCache) error {
	docJSON, err := json.Marshal(cache.Document)
	if err != nil {
		return err
	}

	query := `
		UPDATE oidc_discovery_cache
		SET discovery_doc = $1, cached_at = $2, expires_at = $3
		WHERE id = $4
	`
	_, err = r.db.ExecContext(ctx, query, docJSON, cache.CachedAt, cache.ExpiresAt, cache.ID)
	return err
}

func (r *oidcDiscoveryCacheRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM oidc_discovery_cache WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *oidcDiscoveryCacheRepo) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM oidc_discovery_cache WHERE expires_at < NOW()`
	_, err := r.db.ExecContext(ctx, query)
	return err
}
