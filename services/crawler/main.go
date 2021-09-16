package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jinayshah7/distributedSearchEngine/services/crawler/crawler"
	"github.com/jinayshah7/distributedSearchEngine/services/crawler/partition"
	"github.com/jinayshah7/distributedSearchEngine/services/linkgraph/linkgraphapi"
	linkgraphproto "github.com/jinayshah7/distributedSearchEngine/services/linkgraph/linkgraphapi/proto"
	"github.com/jinayshah7/distributedSearchEngine/services/textindexer/textindexerapi"
	textindexerproto "github.com/jinayshah7/distributedSearchEngine/services/textindexer/textindexerapi/proto"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
)

func main() {
	var wg sync.WaitGroup

	linkGraphURL := os.GetEnv("LINK_GRAPH_URL")
	textIndexerURL := os.GetEnv("TEXT_INDEXER_URL")
	partitionDetectionMode := os.GetEnv("PARITION_DETECTION_MODE")

	partitionDetector, err := getPartitionDetector(appCtx.String("partition-detection-mode"))
	if err != nil {
		return err
	}

	graphAPI, indexerAPI, err := getAPIs(ctx, linkGraphURL, textIndexerURL)
	if err != nil {
		return err
	}

	crawlerConfig := crawler.Config{
		FetchWorkers:      os.GetEnv("NUM_WORKERS"),
		UpdateInterval:    os.GetEnv("UPDATE_INTERVAL"),
		ReIndexThreshold:  os.GetEnv("RE_INDEX_THRESHOLD"),
		GraphAPI:          graphAPI,
		IndexAPI:          indexerAPI,
		PartitionDetector: partitionDetector,
	}
	crawlerService, err := crawler.NewService(crawlerConfig)
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := crawlerService.Run(ctx); err != nil {
			log.Fatalf("err", err).Error("crawler service exited with error")
		}
	}()

	wg.Wait()
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

func getPartitionDetector(mode string) (partition.Detector, error) {
	switch {
	case mode == "single":
		return partition.Fixed{Partition: 0, NumPartitions: 1}, nil
	case strings.HasPrefix(mode, "dns="):
		tokens := strings.Split(mode, "=")
		return partition.DetectFromSRVRecords(tokens[1]), nil
	default:
		return nil, xerrors.Errorf("unsupported partition detection mode: %q", mode)
	}
}
