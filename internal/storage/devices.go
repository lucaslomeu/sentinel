package storage

import (
	"sync"
	"time"

	"github.com/lucaslomeu/sentinel/internal/collector"
)

type DeviceStore struct {
	devices map[string]collector.Device // key: MAC
	mu      sync.RWMutex
}

func NewDeviceStore() *DeviceStore {
	return &DeviceStore{
		devices: make(map[string]collector.Device),
	}
}

func (s *DeviceStore) UpdateDevices(newDevices []collector.Device) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, d := range newDevices {
		s.devices[d.MAC] = d
	}
}

func (s *DeviceStore) GetActiveDevices(timeout time.Duration) []collector.Device {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cutoff := time.Now().Add(-timeout)
	result := make([]collector.Device, 0)

	for _, d := range s.devices {
		if d.LastSeen.After(cutoff) {
			result = append(result, d)
		}
	}

	return result
}

func (s *DeviceStore) CleanupStaleDevices(maxAge time.Duration) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	removed := 0

	for mac, d := range s.devices {
		if d.LastSeen.Before(cutoff) {
			delete(s.devices, mac)
			removed++
		}
	}

	return removed
}
