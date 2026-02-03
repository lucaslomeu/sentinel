package collector

import (
	"log"
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
	store.UpdateDevices(devices)
	return nil
}
