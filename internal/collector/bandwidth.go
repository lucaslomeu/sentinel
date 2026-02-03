package collector

import (
	"fmt"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/net"
)

type InterfaceStats struct {
	Name      string
	RxBytes   uint64
	TxBytes   uint64
	Timestamp time.Time
}

func GetInterfaceStats() (*InterfaceStats, error) {
    counters, err := net.IOCounters(true) 
    if err != nil {
        return nil, err
    }

	iface, err := getDefaultInterface()
	if err != nil {
		return nil, err
	}
    
    for _, counter := range counters {
        if counter.Name == iface {
            return &InterfaceStats{
                Name:      counter.Name,
                RxBytes:   counter.BytesRecv,
                TxBytes:   counter.BytesSent,
                Timestamp: time.Now(),
            }, nil
        }
    }
    
    return nil, fmt.Errorf("Interface %s not found", iface)
}

func getDefaultInterface() (string, error) {
    counters, err := net.IOCounters(true)
    if err != nil {
        return "", err
    }
    
    var maxBytes uint64
    var defaultIface string
    
    for _, c := range counters {
        if isIgnoredInterface(c.Name) {
            continue
        }
        
        total := c.BytesRecv + c.BytesSent
        if total > maxBytes {
            maxBytes = total
            defaultIface = c.Name
        }
    }
    
    if defaultIface == "" {
        return "", fmt.Errorf("No valid network interface found")
    }
    
    fmt.Printf("Detected primary interface: %s\n", defaultIface)
    return defaultIface, nil
}

func isIgnoredInterface(name string) bool {
    ignored := []string{
        "lo",      // Linux loopback
        "lo0",     // macOS loopback
        "awdl0",   // Apple Wireless Direct Link
        "llw0",    // Apple low latency
        "utun",    // VPN tunnels (prefix)
        "bridge",  // Docker/VM bridges
    }
    
    for _, prefix := range ignored {
        if strings.HasPrefix(name, prefix) {
            return true
        }
    }
    
    return false
}

func CalculateBandwidth(old, new *InterfaceStats) (rxMbps, txMbps float64) {
    duration := new.Timestamp.Sub(old.Timestamp).Seconds()
    if duration == 0 {
        return 0, 0
    }
    
    rxMbps = float64(new.RxBytes-old.RxBytes) * 8 / duration / 1_000_000
    txMbps = float64(new.TxBytes-old.TxBytes) * 8 / duration / 1_000_000
    
    return
}