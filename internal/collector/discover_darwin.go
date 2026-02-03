package collector

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

func DiscoverDevices() ([]Device, error) {
	arpOut, err := run("arp", "-a")
	if err != nil {
		return nil, fmt.Errorf("arp -a failed: %w", err)
	}

	ndpOut, err := run("ndp", "-a")
	if err != nil {
		ndpOut = ""
	}

	devs := make([]Device, 0, 64)
	seen := map[string]bool{}

	now := time.Now()
	for _, d := range parseArpDarwin(arpOut, now) {
		if !seen[d.MAC] {
			seen[d.MAC] = true
			devs = append(devs, d)
		}
	}
	for _, d := range parseNdpDarwin(ndpOut, now) {
		if !seen[d.MAC] {
			seen[d.MAC] = true
			devs = append(devs, d)
		}
	}

	return devs, nil
}

func run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, string(out))
	}
	return string(out), nil
}

// arp -a example: "? (192.168.1.1) at aa:bb:cc:dd:ee:ff on en0 ifscope [ethernet]"
var arpRe = regexp.MustCompile(`\((\d+\.\d+\.\d+\.\d+)\)\s+at\s+([0-9a-fA-F:]+)\b`)

func parseArpDarwin(s string, now time.Time) []Device {
	sc := bufio.NewScanner(strings.NewReader(s))
	out := make([]Device, 0)

	for sc.Scan() {
		line := sc.Text()
		m := arpRe.FindStringSubmatch(line)
		if len(m) < 3 {
			continue
		}

		ip := m[1]
		mac := normalizeMac(m[2])

		if ip == "" || mac == "" || isInvalidMac(mac) {
			continue
		}

		out = append(out, Device{
			IP:       ip,
			MAC:      mac,
			LastSeen: now,
		})
	}
	return out
}

// checks if an IPv6 address is link-local (fe80::/10)
func isLinkLocalIPv6(addr string) bool {
	addr = strings.ToLower(addr)
	return strings.HasPrefix(addr, "fe8") ||
		strings.HasPrefix(addr, "fe9") ||
		strings.HasPrefix(addr, "fea") ||
		strings.HasPrefix(addr, "feb")
}

// ndp -a example "fe80::a:b:c:d%en0  0:11:22:33:44:55  0  0  en0  R"
var ndpRe = regexp.MustCompile(`^([0-9a-fA-F:]+)(?:%[^\s]+)?\s+([0-9a-fA-F:]+)\b`)

func parseNdpDarwin(s string, now time.Time) []Device {
	if strings.TrimSpace(s) == "" {
		return nil
	}

	sc := bufio.NewScanner(bytes.NewBufferString(s))
	out := make([]Device, 0)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(strings.ToLower(line), "neighbor") {
			continue
		}

		m := ndpRe.FindStringSubmatch(line)
		if len(m) < 3 {
			continue
		}

		ipv6 := m[1]
		mac := normalizeMac(m[2])

		if mac == "" || isInvalidMac(mac) || !isLinkLocalIPv6(ipv6) {
			continue
		}

		out = append(out, Device{
			IP:       ipv6,
			MAC:      mac,
			LastSeen: now,
		})
	}
	return out
}

// convert MAC address to standard format (aa:bb:cc:dd:ee:ff)
func normalizeMac(mac string) string {
	mac = strings.ToLower(strings.ReplaceAll(mac, "-", ":"))

	parts := strings.Split(mac, ":")
	if len(parts) == 6 {
		for i, part := range parts {
			if len(part) == 1 {
				parts[i] = "0" + part
			}
		}
		mac = strings.Join(parts, ":")
	}

	return mac
}

func isInvalidMac(mac string) bool {
	if len(mac) != 17 {
		return true
	}

	if mac == "00:00:00:00:00:00" || mac == "ff:ff:ff:ff:ff:ff" {
		return true
	}

	if strings.HasPrefix(mac, "01:") || strings.HasPrefix(mac, "33:") {
		return true
	}

	return false
}
