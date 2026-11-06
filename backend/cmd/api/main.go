package main

import (
	"context"
	"log"
	"time"

	delivery "stockflow/internal/delivery/http"
	"stockflow/internal/infrastructure/config"
	"stockflow/internal/infrastructure/postgres"
	"stockflow/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	healthService := service.NewHealthService(postgres.NewHealthRepository(pool))
	if err := healthService.CheckDatabase(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("database connection ok")

	router := delivery.NewRouter()
	if err := router.Start(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
