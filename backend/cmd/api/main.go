package main

import (
	"context"
	"log"
	"time"

	delivery "stockflow/internal/delivery/http"
	"stockflow/internal/delivery/http/handler"
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

	repos := postgres.NewRepositories(pool)
	txRunner := postgres.NewTxRunner(pool)

	productHandler := handler.NewProductHandler(service.NewProductService(repos.Products))
	stockHandler := handler.NewStockHandler(service.NewStockService(repos, txRunner))
	orderHandler := handler.NewOrderHandler(service.NewOrderService(repos, txRunner))

	router := delivery.NewRouter(productHandler, stockHandler, orderHandler)
	if err := router.Start(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
