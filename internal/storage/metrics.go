package storage

import (
	"sync"

	"github.com/lucaslomeu/sentinel/internal/collector"
)

type MetricsStore struct {
	mu        sync.RWMutex
	lastStats *collector.InterfaceStats
}

func NewMetricsStore() *MetricsStore {
	return &MetricsStore{}
}

func (m *MetricsStore) SetLastStats(stats *collector.InterfaceStats) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastStats = stats
}

func (m *MetricsStore) GetLastStats() *collector.InterfaceStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastStats
}
