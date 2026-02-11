package collector

import "time"

type Device struct {
	IP       string    `json:"ip"`
	MAC      string    `json:"mac"`
	Latency  *int      `json:"latency_ms,omitempty"`
	LastSeen time.Time `json:"last_seen"`
}

type DeviceResponse struct {
	IP        string `json:"ip"`
	MAC       string `json:"mac"`
	LatencyMs *int   `json:"latency_ms,omitempty"`
	LastSeen  string `json:"last_seen"`
}

type DevicesResponse struct {
	Devices []DeviceResponse `json:"devices"`
	Count   int              `json:"count"`
	Updated string           `json:"updated"`
}

type BandwidthResponse struct {
	Interface        string  `json:"interface"`
	RxMbps           float64 `json:"rx_mbps"`
	TxMbps           float64 `json:"tx_mbps"`
	Timestamp        string  `json:"timestamp"`
	FirstMeasurement bool    `json:"first_measurement"`
}
