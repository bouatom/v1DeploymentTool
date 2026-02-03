package scanner

import (
	"context"
	"net"
	"strconv"
	"time"
)

type Probe interface {
	Probe(ctx context.Context, host string) (ProbeResult, error)
}

type ProbeResult struct {
	Reachable bool
	OpenPorts []int
}

type PortProbe struct {
	Ports   []int
	Timeout time.Duration
}

func (probe PortProbe) Probe(ctx context.Context, host string) (ProbeResult, error) {
	if len(probe.Ports) == 0 {
		return ProbeResult{}, nil
	}

	timeout := probe.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	dialer := &net.Dialer{
		Timeout: timeout,
	}

	var openPorts []int
	for _, port := range probe.Ports {
		address := net.JoinHostPort(host, strconv.Itoa(port))
		conn, err := dialer.DialContext(ctx, "tcp", address)
		if err != nil {
			continue
		}
		_ = conn.Close()
		openPorts = append(openPorts, port)
	}

	return ProbeResult{
		Reachable: len(openPorts) > 0,
		OpenPorts: openPorts,
	}, nil
}
