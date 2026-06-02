package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	// Replace with your exact module name from go.mod
	"github.com/shashisharma307703/vedantam/config"
)

// InitPool parses PoolConfig and initializes a finely-tuned pgxpool connection pool
func InitPool(ctx context.Context, cfg config.PoolConfig) (*pgxpool.Pool, error) {
	// 1. Build standard connection string uri structure safely
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&connect_timeout=%d",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		int(cfg.ConnectionTimeout.Seconds()),
	)

	// 2. Parse DSN baseline into native connection configurations structure object
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection configuration: %w", err)
	}

	// 3. Bind your dynamic custom pool parameters explicitly onto the poolConfig object
	poolCfg.MaxConns = cfg.MaxConnections
	poolCfg.MinConns = cfg.MinConnections
	poolCfg.MaxConnLifetime = cfg.ConnMaxLifetime
	poolCfg.MaxConnIdleTime = cfg.ConnMaxIdleTime
	poolCfg.HealthCheckPeriod = 1 * time.Minute // System-level periodic verification runtime line

	// 4. Configure internal underlying connection timeout options
	poolCfg.ConnConfig.ConnectTimeout = cfg.ConnectionTimeout

	// 5. Establish connection context pool with custom configurations
	connectCtx, cancel := context.WithTimeout(ctx, cfg.AcquireTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to construct pgx connection pool: %w", err)
	}

	// 6. Execute a Ping request to verify the pool is ready and network boundaries are safe
	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database connectivity check ping failed: %w", err)
	}

	return pool, nil
}
