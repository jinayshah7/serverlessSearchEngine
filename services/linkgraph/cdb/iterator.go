package cdb

import (
	"database/sql"

	"github.com/jinayshah7/distributedSearchEngine/services/linkgraph/graph"
	"golang.org/x/xerrors"
)

type linkIterator struct {
	rows        *sql.Rows
	lastErr     error
	latchedLink *graph.Link
}

func (i *linkIterator) Next() bool {
	if i.lastErr != nil || !i.rows.Next() {
		return false
	}

	l := new(graph.Link)
	i.lastErr = i.rows.Scan(&l.ID, &l.URL, &l.RetrievedAt)
	if i.lastErr != nil {
		return false
	}
	l.RetrievedAt = l.RetrievedAt.UTC()

	i.latchedLink = l
	return true
}

func (i *linkIterator) Error() error {
	return i.lastErr
}

func (i *linkIterator) Close() error {
	err := i.rows.Close()
	if err != nil {
		return xerrors.Errorf("link iterator: %w", err)
	}
	return nil
}

func (i *linkIterator) Link() *graph.Link {
	return i.latchedLink
}

type edgeIterator struct {
	rows        *sql.Rows
	lastErr     error
	latchedEdge *graph.Edge
}

func (i *edgeIterator) Next() bool {
	if i.lastErr != nil || !i.rows.Next() {
		return false
	}

	e := new(graph.Edge)
	i.lastErr = i.rows.Scan(&e.ID, &e.Src, &e.Dst, &e.UpdatedAt)
	if i.lastErr != nil {
		return false
	}
	e.UpdatedAt = e.UpdatedAt.UTC()

	i.latchedEdge = e
	return true
}

func (i *edgeIterator) Error() error {
	return i.lastErr
}

func (i *edgeIterator) Close() error {
	err := i.rows.Close()
	if err != nil {
		return xerrors.Errorf("edge iterator: %w", err)
	}
	return nil
}

func (i *edgeIterator) Edge() *graph.Edge {
	return i.latchedEdge
}
