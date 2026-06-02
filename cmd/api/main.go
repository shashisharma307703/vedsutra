package main

import (
	"context"
	"log"

	"github.com/shashisharma307703/vedantam/config"
	"github.com/shashisharma307703/vedantam/internal/app"
)

func main() {
	ctx := context.Background()

	// Parse clean configuration variables
	cfg := config.Load()

	// Build runtime framework modules
	application, err := app.NewApp(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize application engine: %v", err)
	}
	defer application.Close()

	// Boot up application execution listener context loop
	if err := application.Run(); err != nil {
		log.Fatalf("server crash boundary encountered: %v", err)
	}
}
