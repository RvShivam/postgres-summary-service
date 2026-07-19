package domain

import (
	"time"

	"github.com/google/uuid"
)

type SummaryOverview struct {
	ID                uuid.UUID
	ExternalSummaryID string
	Host              string
	Port              int
	User              string
	DBName            string
	SyncedAt          time.Time
}
