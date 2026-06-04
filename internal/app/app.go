package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shashisharma307703/vedantam/config"
	"github.com/shashisharma307703/vedantam/internal/handler"
	"github.com/shashisharma307703/vedantam/internal/logger"
	"github.com/shashisharma307703/vedantam/internal/repository"
	"github.com/shashisharma307703/vedantam/internal/service"
)

type App struct {
	Cfg    *config.Config
	Pool   *pgxpool.Pool
	Server *Server
	Logger logger.Logger
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	// 1. Initialize the Custom Connection Pool using repository.InitPool
	pool, err := repository.InitPool(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("database initialization failure: %w", err)
	}

	// 2. Initialize logger
	appLogger := logger.NewLogger("vedsutra")

	// 3. Instantiate Architecture Layers
	repo := repository.NewRepository(pool)

	orgSvc := service.NewOrgService(repo)
	orgHnd := handler.NewOrgHandler(orgSvc)

	classSvc := service.NewClassService(repo)
	classHnd := handler.NewClassHandler(classSvc)

	// 4. Initialize auth services
	// Create SQL adapter for pgxpool
	sqlDB := &sql.DB{} // Placeholder - pgxpool doesn't directly implement sql.DB
	// TODO: Either refactor repositories to use pgxpool directly or create a proper adapter
	
	authSessionRepo := repository.NewAuthSessionRepository(sqlDB)
	authProviderRepo := repository.NewAuthProviderRepository(sqlDB)
	oidcCacheRepo := repository.NewOIDCDiscoveryCacheRepository(sqlDB)
	userRepo := repository.NewUserRepository(sqlDB)

	tokenService := service.NewTokenService(&cfg.Auth)
	oidcDiscoveryService := service.NewOIDCDiscoveryService(&cfg.Auth, oidcCacheRepo)
	
	authService := service.NewAuthService(
		&cfg.Auth,
		tokenService,
		oidcDiscoveryService,
		authSessionRepo,
		authProviderRepo,
		userRepo,
	)

	authHandler := handler.NewAuthHandler(authService, appLogger)

	// 5. Instantiate the Server Module
	srv := NewServer(
		cfg.Server.Port,
		cfg.Server.ReadTimeout,
		orgHnd,
		classHnd,
		authService,
		authHandler,
		appLogger,
	)

	return &App{
		Cfg:    cfg,
		Pool:   pool,
		Server: srv,
		Logger: appLogger,
	}, nil
}

func (a *App) Run() error {
	a.Logger.Infof("Server booting on %s...", a.Cfg.Server.Port)
	return a.Server.Start()
}

func (a *App) Close() {
	if a.Pool != nil {
		a.Pool.Close()
		a.Logger.Info("Database connection pool closed successfully.")
	}
}
