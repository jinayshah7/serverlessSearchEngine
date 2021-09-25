package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"strings"
	"sync"

	"google.golang.org/grpc"

func main() error {
	var wg sync.WaitGroup
	_, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	grpcPort := os.Getenv("GRPC_PORT")
	documentRepositoryURL := os.Getenv("DOCUMENT_REPOSITORY_URL")

	indexer, err := getTextIndexer(documentRepositoryURL)
	if err != nil {
		return err
	}

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort)))
	if err != nil {
		return err
	}
	defer func() { _ = grpcListener.Close() }()

	wg.Add(1)
	go func() {
		defer wg.Done()
		grpcServer := grpc.NewServer()
		documentRepositoryServer : = NewDocumentRepositoryServer(indexer))
		documentRepository.RegisterDocumentRepositoryServer(grpcServer, documentRepositoryServer)
		_ = srv.Serve(grpcListener)
	}()

	wg.Wait()
	return nil
}

func getDocumentRepository(documentRepositoryURL string) (documentRepository.documentRepositoryServer, error) {
	if documentRepositoryURL == "" {
		return nil, errors.New(fmt.Sprintf("document repository url not found"))
	}

	url, err := url.Parse(documentRepositoryURL)
	if err != nil {
		return nil, errors.New(Sprintf("could not parse document repository URL: %w", err))
	}

	nodes := strings.Split(uri.Host, ",")
	for i := 0; i < len(nodes); i++ {
		nodes[i] = "http://" + nodes[i]
	}
	return es.NewElasticSearchIndexer(nodes, false)

}
