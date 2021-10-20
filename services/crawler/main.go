package main

import (
	"context"
	"errors"
	"time"

	"github.com/jinayshah7/distributedSearchEngine/services/linkgraph/linkgraphapi"
	linkgraphproto "github.com/jinayshah7/distributedSearchEngine/services/linkgraph/linkgraphapi/proto"
	"github.com/jinayshah7/distributedSearchEngine/services/textindexer/textindexerapi"
	textindexerproto "github.com/jinayshah7/distributedSearchEngine/services/textindexer/textindexerapi/proto"
	"google.golang.org/grpc"
)

func main() {
}

func getAPIs(ctx context.Context, linkGraphURL, textIndexerURL string) (*linkgraphapi.LinkGraphClient, *textindexerapi.TextIndexerClient, error) {

	if linkGraphURL == "" {
		return nil, nil, errors.New("link graph API must be specified with --link-graph-api")
	}

	if textIndexerURL == "" {
		return nil, nil, errors.New("text indexer API must be specified with --text-indexer-api")
	}

	dialCtx, cancelFn := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFn()

	linkGraphConnection, err := grpc.DialContext(dialCtx, linkGraphURL, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, nil, errors.New("could not connect to link graph API: %w", err)
	}

	graphClient := linkgraphapi.NewLinkGraphClient(ctx, linkgraphproto.NewLinkGraphClient(linkGraphConnection))

	dialCtx, cancelFn = context.WithTimeout(ctx, 5*time.Second)
	defer cancelFn()

	indexerConnection, err := grpc.DialContext(dialCtx, textIndexerURL, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, nil, errors.New("could not connect to text indexer API: %w", err)
	}

	indexerClient := textindexerapi.NewTextIndexerClient(ctx, textindexerproto.NewTextIndexerClient(indexerConnection))

	return graphClient, indexerClient, nil
}
