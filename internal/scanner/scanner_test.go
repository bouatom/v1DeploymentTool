package scanner

import (
	"context"
	"testing"

	"v1-sg-deployment-tool/internal/targets"
)

type fakeProbe struct{}

func (probe fakeProbe) Probe(ctx context.Context, host string) (ProbeResult, error) {
	return ProbeResult{
		Reachable: true,
		OpenPorts: []int{22},
	}, nil
}

func TestScanTargetsReturnsResults(t *testing.T) {
	specs, errs := targets.ParseInputs([]string{"127.0.0.1", "localhost"})
	if len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}

	results, err := ScanTargets(context.Background(), specs, ScannerConfig{
		MaxConcurrency: 2,
		RatePerSecond:  10,
	}, fakeProbe{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}
