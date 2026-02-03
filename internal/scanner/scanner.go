package scanner

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"v1-sg-deployment-tool/internal/targets"
)

type ScanResult struct {
	Host      string
	Source    string
	Reachable bool
	OpenPorts []int
	Error     string
}

func ScanTargets(ctx context.Context, specs []targets.TargetSpec, config ScannerConfig, probe Probe) ([]ScanResult, error) {
	if len(specs) == 0 {
		return nil, errors.New("no targets to scan")
	}
	if probe == nil {
		return nil, errors.New("probe is required")
	}

	normalized := NormalizeConfig(config)
	hosts := expandSpecs(specs)

	results := make([]ScanResult, 0, len(hosts))
	resultsChan := make(chan ScanResult, len(hosts))
	jobs := make(chan scanJob)

	var wg sync.WaitGroup
	for i := 0; i < normalized.MaxConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				resultsChan <- scanHost(ctx, job, normalized, probe)
			}
		}()
	}

	go func() {
		defer close(jobs)
		ticker := time.NewTicker(time.Second / time.Duration(normalized.RatePerSecond))
		defer ticker.Stop()
		for _, job := range hosts {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				jobs <- job
			}
		}
	}()

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for result := range resultsChan {
		results = append(results, result)
	}

	return results, nil
}

type scanJob struct {
	host   string
	source string
}

func scanHost(ctx context.Context, job scanJob, config ScannerConfig, probe Probe) ScanResult {
	probeCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	result, err := probe.Probe(probeCtx, job.host)
	if err != nil {
		return ScanResult{
			Host:   job.host,
			Source: job.source,
			Error:  err.Error(),
		}
	}

	return ScanResult{
		Host:      job.host,
		Source:    job.source,
		Reachable: result.Reachable,
		OpenPorts: result.OpenPorts,
	}
}

func expandSpecs(specs []targets.TargetSpec) []scanJob {
	var jobs []scanJob
	for _, spec := range specs {
		switch spec.Kind {
		case targets.TargetKindHostname:
			jobs = append(jobs, scanJob{
				host:   spec.Hostname,
				source: spec.Original,
			})
		case targets.TargetKindIP:
			jobs = append(jobs, scanJob{
				host:   spec.IP.String(),
				source: spec.Original,
			})
		case targets.TargetKindCIDR:
			for _, ip := range expandCIDR(spec.CIDR) {
				jobs = append(jobs, scanJob{
					host:   ip,
					source: spec.Original,
				})
			}
		default:
			continue
		}
	}

	return jobs
}

func expandCIDR(ipNet *net.IPNet) []string {
	if ipNet == nil {
		return nil
	}

	var ips []string
	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); ip = incrementIP(ip) {
		ips = append(ips, ip.String())
	}

	return ips
}

func incrementIP(ip net.IP) net.IP {
	updated := make(net.IP, len(ip))
	copy(updated, ip)

	for i := len(updated) - 1; i >= 0; i-- {
		updated[i]++
		if updated[i] != 0 {
			break
		}
	}

	return updated
}
