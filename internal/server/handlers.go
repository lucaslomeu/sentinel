package server

import (
	"net/http"
	"time"

	"github.com/lucaslomeu/sentinel/internal/collector"
	"github.com/lucaslomeu/sentinel/internal/storage"
)

type Handlers struct {
	metricsStore *storage.MetricsStore
	deviceStore  *storage.DeviceStore
}

func NewHandlers(ms *storage.MetricsStore, ds *storage.DeviceStore) *Handlers {
	return &Handlers{
		metricsStore: ms,
		deviceStore:  ds,
	}
}

func (h *Handlers) Root(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Sentinel Dashboard - Monitor your network bandwidth and devices"))
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (h *Handlers) Bandwidth(w http.ResponseWriter, r *http.Request) {
	currentStats, err := collector.GetInterfaceStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lastStats := h.metricsStore.GetLastStats()

	var rxMbps, txMbps float64
	first := false

	if lastStats == nil {
		first = true
	} else {
		rxMbps, txMbps = collector.CalculateBandwidth(lastStats, currentStats)
	}

	h.metricsStore.SetLastStats(currentStats)

	resp := collector.BandwidthResponse{
		Interface:        currentStats.Name,
		RxMbps:           rxMbps,
		TxMbps:           txMbps,
		Timestamp:        currentStats.Timestamp.Format(time.RFC3339),
		FirstMeasurement: first,
	}

	WriteJSON(w, http.StatusOK, resp)
}

func (h *Handlers) Devices(w http.ResponseWriter, r *http.Request) {
	devices := h.deviceStore.GetActiveDevices(10 * time.Minute)

	resp := collector.DevicesResponse{
		Devices: toDeviceResponses(devices),
		Count:   len(devices),
		Updated: time.Now().Format(time.RFC3339),
	}

	WriteJSON(w, http.StatusOK, resp)
}

func toDeviceResponse(d collector.Device) collector.DeviceResponse {
	return collector.DeviceResponse{
		IP:       d.IP,
		MAC:      d.MAC,
		LastSeen: d.LastSeen.Format(time.RFC3339),
	}
}

func toDeviceResponses(devices []collector.Device) []collector.DeviceResponse {
	out := make([]collector.DeviceResponse, len(devices))
	for i := range devices {
		out[i] = toDeviceResponse(devices[i])
	}
	return out
}
