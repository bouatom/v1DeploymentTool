package handlers

import (
	"testing"
	"time"

	"v1-sg-deployment-tool/internal/models"
	"v1-sg-deployment-tool/internal/store"
)

func TestPredictSuccessWindows(t *testing.T) {
	reachable := true
	record := store.AssessmentRecord{
		TargetID:  "t1",
		OS:        models.TargetOSWindows,
		Reachable: &reachable,
		OpenPorts: []int{5986},
	}

	score := predictSuccess(record)
	if score < 80 {
		t.Fatalf("expected high score, got %d", score)
	}
}

func TestBuildAssessmentIncludesGuidelines(t *testing.T) {
	reachable := false
	scannedAt := time.Now().UTC()
	record := store.AssessmentRecord{
		TargetID:  "t1",
		Hostname:  "host-1",
		OS:        models.TargetOSLinux,
		Reachable: &reachable,
		ScannedAt: &scannedAt,
	}

	response := buildAssessment(record)
	if response.Label != "host-1" {
		t.Fatalf("expected label to be host-1")
	}
	if len(response.Guidelines) == 0 {
		t.Fatalf("expected guidelines")
	}
}
