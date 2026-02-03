package policy

import (
	"testing"

	"v1-sg-deployment-tool/internal/errors"
)

func TestDecideNextAction(t *testing.T) {
	action := DecideNextAction(errors.CodeAuthDenied, true, true)
	if action != ActionSwitchAuth {
		t.Fatalf("expected switch auth, got %s", action)
	}

	action = DecideNextAction(errors.CodeInstallFailed, false, false)
	if action != ActionAbort {
		t.Fatalf("expected abort, got %s", action)
	}
}
