package collector

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync/atomic"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

var pingSeq uint32

func PingDevice(ip string, timeout time.Duration) (int, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil || parsedIP.To4() == nil {
		return 0, fmt.Errorf("not an IPv4 address: %s", ip)
	}

	// i think this need run go with sudo
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	// ICMP only supports 16 bits
	// 0xffff in hex is 65535 in decimal = max value for a 16-bit unsigned integer
	// pid has a range of 0-65535
	id := os.Getpid() & 0xffff
	seq := int(atomic.AddUint32(&pingSeq, 1) & 0xffff)

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   id,
			Seq:  seq,
			Data: []byte("sentinel-ping"),
		},
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return 0, err
	}

	dst := &net.IPAddr{IP: parsedIP}

	start := time.Now()
	if _, err := conn.WriteTo(msgBytes, dst); err != nil {
		return 0, err
	}

	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return 0, err
	}

	reply := make([]byte, 1500)
	targetIP := parsedIP.To4()

	for {
		n, peer, err := conn.ReadFrom(reply)
		if err != nil {
			return 0, err
		}

		parsed, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), reply[:n])
		if err != nil {
			continue
		}
		if parsed.Type != ipv4.ICMPTypeEchoReply {
			continue
		}

		echo, ok := parsed.Body.(*icmp.Echo)
		if !ok || echo.ID != id || echo.Seq != seq {
			continue
		}

		if peerIP, ok := peer.(*net.IPAddr); ok {
			if ipAddr := peerIP.IP.To4(); ipAddr != nil && !ipAddr.Equal(targetIP) {
				continue
			}
		}

		elapsed := time.Since(start)
		latencyMs := int(elapsed.Milliseconds())
		log.Printf("ping %s latency=%dms (%dus)", ip, latencyMs, elapsed.Microseconds())
		return latencyMs, nil
	}
}
