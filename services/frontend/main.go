package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/jinayshah7/distributedSearchEngine/services/frontend/frontend"
	"github.com/jinayshah7/distributedSearchEngine/services/linkgraph/linkgraphapi"
	linkrepositoryproto "github.com/jinayshah7/distributedSearchEngine/services/linkgraph/linkgraphapi/proto"
	"github.com/jinayshah7/distributedSearchEngine/services/textindexer/textindexerapi"
	documentrepositoryproto "github.com/jinayshah7/distributedSearchEngine/services/textindexer/textindexerapi/proto"
	"google.golang.org/grpc"
)

func main() {
	var wg sync.WaitGroup

	linkRepository, documentRepository, err := getAPIs(ctx, os.GetEnv("LINK_REPOSITORY_URL"), os.GetEnv("DOCUMENT_REPOSITORY_URL"))
	if err != nil {
		log.Fatalf("Failed to fetch linkRepository and documentRepository")
		panic(err)
	}

	frontendCfg := frontend.Config{
		ListenAddr:       fmt.Sprintf(":%d", os.GetEnv("FRONTEND_PORT")),
		ResultsPerPage:   os.GetEnv("RESULTS_PER_PAGE"),
		MaxSummaryLength: os.GetEnv("MAX_SUMMARY_LENGTH"),
		GraphAPI:         linkRepository,
		IndexAPI:         documentRepository,
	}
	frontentService, err := frontend.NewService(frontendConfig)
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := frontentService.Run(ctx); err != nil {
			log.Error("front-end service exited with error %w", err)
			panic(err)
		}
	}()

}

func getAPIs(ctxcontext.Context, linkRepositoryURL, documentRepositoryURL string) (*linkgraphapi.LinkGraphClient, *textindexerapi.TextIndexerClient, error) {
	if linkRepositoryURL == "" {
		return nil, nil, errors.New("link graph URL must be specified with --link-graph-api")
	}
	if documentRepositoryURL == "" {
		return nil, nil, errors.New("text indexer URL must be specified with --text-indexer-api")
	}

	dialCtx, cancelFn := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFn()

	linkRepositoryConnection, err := grpc.DialContext(dialCtx, linkRepositoryURL, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, nil, errors.New("could not connect to link graph API: %w", err)

	}
	linkRepositoryClient := linkgraphapi.NewLinkGraphClient(ctx, linkrepositoryproto.NewLinkGraphClient(linkRepositoryConnection))

	dialCtx, cancelFn = context.WithTimeout(ctx, 5*time.Second)
	defer cancelFn()

	documentRepositoryConnection, err := grpc.DialContext(dialCtx, documentRepositoryURL, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, nil, errors.New("could not connect to text indexer API: %w", err)
	}

	documentRepositoryClient := textindexerapi.NewTextIndexerClient(ctx, documentrepositoryproto.NewTextIndexerClient(documentRepositoryConnection))

	return linkRepositoryClient, documentRepositoryClient, nil
}
