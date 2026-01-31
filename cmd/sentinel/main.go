package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lucaslomeu/sentinel/internal/collector"
	"github.com/lucaslomeu/sentinel/internal/storage"
)

type Response struct {
	Interface 			string  `json:"interface"`
	RxMbps    			float64 `json:"rx_mbps"`
	TxMbps    			float64 `json:"tx_mbps"`
	Timestamp 			string  `json:"timestamp"`
	FirstMeasurement	bool 	`json:"first_measurement"`
}

var metricsStore = storage.NewMetricsStore()

func main() {
	r := chi.NewRouter()
    r.Use(middleware.Logger)
    
    // Health check
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })
    
    // Dashboard 
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Sentinel Dashboard"))
    })

	// Bandwidth API
	r.Get("/api/bandwidth", func(w http.ResponseWriter, r *http.Request) {
		currentStats, err := collector.GetInterfaceStats()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Buscar medição anterior
		lastStats := metricsStore.GetLastStats()

		var rxMbps, txMbps float64
		firstMeasurement := false

		if lastStats == nil {
			firstMeasurement = true
			rxMbps = 0
			txMbps = 0
		} else {
			rxMbps, txMbps = collector.CalculateBandwidth(lastStats, currentStats)
		}

		metricsStore.SetLastStats(currentStats)

		fmt.Printf("[%s] RxMbps: %.2f Mbps, TxMbps: %.2f Mbps (First Measurement: %v)\n",
			currentStats.Name, rxMbps, txMbps, firstMeasurement)

		response := Response{
			Interface: currentStats.Name,
			RxMbps:    rxMbps,
			TxMbps:    txMbps,
			Timestamp: currentStats.Timestamp.Format(time.RFC3339),
			FirstMeasurement: firstMeasurement,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

    fmt.Println("Sentinel running on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}