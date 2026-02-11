package collector

import (
	"log"
	"sync"
	"time"
)

type DeviceStore interface {
	UpdateDevices([]Device)
	CleanupStaleDevices(time.Duration) int
}

func StartDeviceScanner(store DeviceStore, interval time.Duration) {
	if err := scanAndUpdate(store); err != nil {
		log.Printf("initial device scan failed: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if err := scanAndUpdate(store); err != nil {
			log.Printf("device scan failed: %v", err)
		}

		removed := store.CleanupStaleDevices(1 * time.Hour)
		if removed > 0 {
			log.Printf("cleaned up %d stale devices", removed)
		}
	}
}

func scanAndUpdate(store DeviceStore) error {
	devices, err := DiscoverDevices()
	if err != nil {
		return err
	}

	const (
		maxConcurrentPings = 16
		pingTimeout        = 2 * time.Second
	)
	sem := make(chan struct{}, maxConcurrentPings)
	var wg sync.WaitGroup

	for i := range devices {
		i := i
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			latency, err := PingDevice(devices[i].IP, pingTimeout)
			if err != nil {
				return
			}

			devices[i].Latency = &latency
		}()
	}

	wg.Wait()
	store.UpdateDevices(devices)
	return nil
}
