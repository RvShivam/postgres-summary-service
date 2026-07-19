package domain

import "github.com/google/uuid"

type Table struct {
	ID       uuid.UUID
	SchemaID uuid.UUID

	Name     string
	RowCount int64
	SizeMB   float64
}
