package linkgraphapi

import (
	"context"
	"io"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"github.com/jinayshah7/distributedSearchEngine/services/linkgraph/graph"
	"github.com/jinayshah7/distributedSearchEngine/services/linkgraph/linkgraphapi/proto"
)

type LinkGraphClient struct {
	ctx    context.Context
	client proto.LinkGraphClient
}

func NewLinkGraphClient(ctx context.Context, rpcClient proto.LinkGraphClient) *LinkGraphClient {
	return &LinkGraphClient{ctx: ctx, client: rpcClient}
}

func (c *LinkGraphClient) UpsertLink(link *graph.Link) error {
	req := &proto.Link{
		Uuid:        link.ID[:],
		Url:         link.URL,
		RetrievedAt: timeToProto(link.RetrievedAt),
	}
	res, err := c.client.UpsertLink(c.ctx, req)
	if err != nil {
		return err
	}

	link.ID = uuidFromBytes(res.Uuid)
	link.URL = res.Url
	if link.RetrievedAt, err = ptypes.Timestamp(res.RetrievedAt); err != nil {
		return err
	}

	return nil
}

func (c *LinkGraphClient) UpsertEdge(edge *graph.Edge) error {
	req := &proto.Edge{
		Uuid:    edge.ID[:],
		SrcUuid: edge.Src[:],
		DstUuid: edge.Dst[:],
	}
	res, err := c.client.UpsertEdge(c.ctx, req)
	if err != nil {
		return err
	}

	edge.ID = uuidFromBytes(res.Uuid)
	if edge.UpdatedAt, err = ptypes.Timestamp(res.UpdatedAt); err != nil {
		return err
	}

	return nil
}

func (c *LinkGraphClient) Links(fromID, toID uuid.UUID, accessedBefore time.Time) (graph.LinkIterator, error) {
	filter, err := ptypes.TimestampProto(accessedBefore)
	if err != nil {
		return nil, err
	}

	req := &proto.Range{
		FromUuid: fromID[:],
		ToUuid:   toID[:],
		Filter:   filter,
	}

	ctx, cancelFn := context.WithCancel(c.ctx)
	stream, err := c.client.Links(ctx, req)
	if err != nil {
		cancelFn()
		return nil, err
	}

	return &linkIterator{stream: stream, cancelFn: cancelFn}, nil
}

func (c *LinkGraphClient) Edges(fromID, toID uuid.UUID, updatedBefore time.Time) (graph.EdgeIterator, error) {
	filter, err := ptypes.TimestampProto(updatedBefore)
	if err != nil {
		return nil, err
	}

	req := &proto.Range{
		FromUuid: fromID[:],
		ToUuid:   toID[:],
		Filter:   filter,
	}

	ctx, cancelFn := context.WithCancel(c.ctx)
	stream, err := c.client.Edges(ctx, req)
	if err != nil {
		cancelFn()
		return nil, err
	}

	return &edgeIterator{stream: stream, cancelFn: cancelFn}, nil
}

func (c *LinkGraphClient) RemoveStaleEdges(from uuid.UUID, updatedBefore time.Time) error {
	req := &proto.RemoveStaleEdgesQuery{
		FromUuid:      from[:],
		UpdatedBefore: timeToProto(updatedBefore),
	}

	_, err := c.client.RemoveStaleEdges(c.ctx, req)
	return err
}

type linkIterator struct {
	stream  proto.LinkGraph_LinksClient
	next    *graph.Link
	lastErr error

	// A function to cancel the context used to perform the streaming RPC. It
	// allows us to abort server-streaming calls from the client side.
	cancelFn func()
}

func (it *linkIterator) Next() bool {
	res, err := it.stream.Recv()
	if err != nil {
		if err != io.EOF {
			it.lastErr = err
		}
		it.cancelFn()
		return false
	}

	lastAccessed, err := ptypes.Timestamp(res.RetrievedAt)
	if err != nil {
		it.lastErr = err
		it.cancelFn()
		return false
	}

	it.next = &graph.Link{
		ID:          uuidFromBytes(res.Uuid),
		URL:         res.Url,
		RetrievedAt: lastAccessed,
	}
	return true
}

func (it *linkIterator) Error() error { return it.lastErr }

func (it *linkIterator) Link() *graph.Link { return it.next }

func (it *linkIterator) Close() error {
	it.cancelFn()
	return nil
}

type edgeIterator struct {
	stream  proto.LinkGraph_EdgesClient
	next    *graph.Edge
	lastErr error

	// A function to cancel the context used to perform the streaming RPC. It
	// allows us to abort server-streaming calls from the client side.
	cancelFn func()
}

func (it *edgeIterator) Next() bool {
	res, err := it.stream.Recv()
	if err != nil {
		if err != io.EOF {
			it.lastErr = err
		}
		it.cancelFn()
		return false
	}

	updatedAt, err := ptypes.Timestamp(res.UpdatedAt)
	if err != nil {
		it.lastErr = err
		it.cancelFn()
		return false
	}

	it.next = &graph.Edge{
		ID:        uuidFromBytes(res.Uuid),
		Src:       uuidFromBytes(res.SrcUuid),
		Dst:       uuidFromBytes(res.DstUuid),
		UpdatedAt: updatedAt,
	}
	return true
}

func (it *edgeIterator) Error() error { return it.lastErr }

func (it *edgeIterator) Edge() *graph.Edge { return it.next }

func (it *edgeIterator) Close() error {
	it.cancelFn()
	return nil
}
