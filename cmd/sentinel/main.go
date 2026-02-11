package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/lucaslomeu/sentinel/internal/collector"
	"github.com/lucaslomeu/sentinel/internal/server"
	"github.com/lucaslomeu/sentinel/internal/storage"
)

func main() {
	metricsStore := storage.NewMetricsStore()
	deviceStore := storage.NewDeviceStore()
	handlers := server.NewHandlers(metricsStore, deviceStore)

	go collector.StartDeviceScanner(deviceStore, 30*time.Second)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", handlers.Root)
	r.Get("/health", handlers.Health)
	r.Get("/api/bandwidth", handlers.Bandwidth)
	r.Get("/api/devices", handlers.Devices)

	log.Println("Sentinel running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}