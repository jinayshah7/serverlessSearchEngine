package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jinayshah7/distributedSearchEngine/proto/linkRepository"
	"github.com/lib/pq"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

var (
	saveLinkQuery = `
INSERT INTO links (url, retrieved_at) VALUES ($1, $2) 
ON CONFLICT (url) DO UPDATE SET retrieved_at=GREATEST(links.retrieved_at, $2)
RETURNING id, retrieved_at
`
	findLinkQuery = "SELECT url, retrieved_at FROM links WHERE id=$1"

	saveEdgeQuery = `
INSERT INTO edges (src, dst, updated_at) VALUES ($1, $2, NOW())
ON CONFLICT (src,dst) DO UPDATE SET updated_at=NOW()
RETURNING id, updated_at
`
	removeOldEdgesQuery = "DELETE FROM edges WHERE src=$1 AND updated_at < $2"

	_ linkRepository.LinkRepositoryServer = (*CockroachDBGraph)(nil)
)

type CockroachDBGraph struct {
	db *sql.DB
}

func NewCockroachDbGraph(dsn string) (*CockroachDBGraph, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	return &CockroachDBGraph{db: db}, nil
}

func (c *CockroachDBGraph) Close() error {
	return c.db.Close()
}

func (c *CockroachDBGraph) SaveLink(ctx context.Context, link *linkRepository.Link) (*linkRepository.Link, error) {
	row := c.db.QueryRow(saveLinkQuery, link.Url, link.RetrievedAt)
	if err := row.Scan(&link.Uuid, &link.RetrievedAt); err != nil {
		return &linkRepository.Link{}, errors.New(fmt.Sprintf("upsert link: %w", err))
	}

	return link, nil
}

func (c *CockroachDBGraph) RemoveOldEdges(ctx context.Context, r *linkRepository.RemoveOldEdgesQuery) (*emptypb.Empty, error) {

	return &emptypb.Empty{}, nil
}

func (c *CockroachDBGraph) FindLink(id uuid.UUID) (*linkRepository.Link, error) {
	row := c.db.QueryRow(findLinkQuery, id)
	link := &linkRepository.Link{Uuid: id[:]}
	if err := row.Scan(&link.Url, &link.RetrievedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(fmt.Sprintf("Unable to find the link"))
		}

		return nil, errors.New(fmt.Sprintf("find link: %w", err))
	}

	return link, nil
}

func (c *CockroachDBGraph) SaveEdge(ctx context.Context, edge *linkRepository.Edge) (*linkRepository.Edge, error) {
	row := c.db.QueryRow(saveEdgeQuery, edge.SourceUuid, edge.DestinationUuid)
	if err := row.Scan(&edge.Uuid, &edge.UpdatedAt); err != nil {
		if isForeignKeyViolationError(err) {
			err = errors.New("Cannot find the edge")
		}
		return &linkRepository.Edge{}, errors.New(fmt.Sprintf("Save edge: %w", err))
	}

	return edge, nil
}

func isForeignKeyViolationError(err error) bool {
	pqErr, valid := err.(*pq.Error)
	if !valid {
		return false
	}

	return pqErr.Code.Name() == "foreign_key_violation"
}
