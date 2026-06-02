package app

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shashisharma307703/vedantam/config"
	"github.com/shashisharma307703/vedantam/internal/handler"
	"github.com/shashisharma307703/vedantam/internal/repository"
	"github.com/shashisharma307703/vedantam/internal/service"
)

type App struct {
	Cfg    *config.Config
	Pool   *pgxpool.Pool
	Server *Server
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	// 1. Initialize the Custom Connection Pool using repository.InitPool
	pool, err := repository.InitPool(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("database initialization failure: %w", err)
	}

	// 2. Instantiate Architecture Layers
	repo := repository.NewRepository(pool)

	orgSvc := service.NewOrgService(repo)
	orgHnd := handler.NewOrgHandler(orgSvc)

	classSvc := service.NewClassService(repo)
	classHnd := handler.NewClassHandler(classSvc)

	// 3. Instantiate the Server Module
	srv := NewServer(cfg.Server.Port, cfg.Server.ReadTimeout, orgHnd, classHnd)

	return &App{
		Cfg:    cfg,
		Pool:   pool,
		Server: srv,
	}, nil
}

func (a *App) Run() error {
	log.Printf("Server booting successfully on configuration location target %s...", a.Cfg.Server.Port)
	return a.Server.Start()
}

func (a *App) Close() {
	if a.Pool != nil {
		a.Pool.Close()
		log.Println("Database connection pool closed successfully.")
	}
}
