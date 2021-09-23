package graph

import (
	"time"

	"github.com/google/uuid"
)

type Link struct {
	ID          uuid.UUID
	URL         string
	RetrievedAt time.Time
}

type Edge struct {
	ID          uuid.UUID
	Source      uuid.UUID
	Destination uuid.UUID

	UpdatedAt time.Time
}

type Graph interface {
	SaveLink(link *Link) error
	SaveEdge(edge *Edge) error
	RemoveOldEdges(fromID uuid.UUID, updatedBefore time.Time) error
}
