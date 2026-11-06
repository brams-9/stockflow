package main

import (
	"log"
	"os"

	delivery "stockflow/internal/delivery/http"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := delivery.NewRouter()
	if err := router.Start(":" + port); err != nil {
		log.Fatal(err)
	}
}
