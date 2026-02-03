package maintenance

import (
	"log"
	"time"
)

type RetentionStore interface {
	DeleteDeploymentsBefore(cutoff time.Time) (int64, error)
	DeleteRunsBefore(cutoff time.Time) (int64, error)
	DeleteAuditLogsBefore(cutoff time.Time) (int64, error)
}

func StartRetentionLoop(store RetentionStore, retentionDays int, logger *log.Logger) {
	if retentionDays <= 0 {
		return
	}

	runCleanup(store, retentionDays, logger)
	ticker := time.NewTicker(24 * time.Hour)

	go func() {
		for range ticker.C {
			runCleanup(store, retentionDays, logger)
		}
	}()
}

func runCleanup(store RetentionStore, retentionDays int, logger *log.Logger) {
	if store == nil {
		return
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -retentionDays)
	deploymentsDeleted, err := store.DeleteDeploymentsBefore(cutoff)
	if err != nil && logger != nil {
		logger.Printf("retention cleanup deployments failed: %v", err)
	}
	runsDeleted, err := store.DeleteRunsBefore(cutoff)
	if err != nil && logger != nil {
		logger.Printf("retention cleanup runs failed: %v", err)
	}
	auditsDeleted, err := store.DeleteAuditLogsBefore(cutoff)
	if err != nil && logger != nil {
		logger.Printf("retention cleanup audit logs failed: %v", err)
	}

	if logger != nil {
		logger.Printf("retention cleanup done: deployments=%d runs=%d audits=%d", deploymentsDeleted, runsDeleted, auditsDeleted)
	}
}
