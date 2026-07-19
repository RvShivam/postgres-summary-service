package domain

import "github.com/google/uuid"

type Schema struct {
	ID        uuid.UUID
	SummaryID uuid.UUID

	Name string

	Tables []Table
}