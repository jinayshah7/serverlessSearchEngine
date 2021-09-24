package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/jinayshah7/distributedSearchEngine/proto/linkRepository"
	"google.golang.org/grpc"
)

func main() {
	var wg sync.WaitGroup
	grpcPort := os.Getenv("GRPC_PORT")
	linkGraphURI := os.Getenv("LINK_GRAPH_URI")

	graph, err := getLinkGraph(linkGraphURI)
	if err != nil {
		panic(err)
	}

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		panic(err)
	}
	defer func() { _ = grpcListener.Close() }()

	wg.Add(1)
	go func() {
		defer wg.Done()
		grpcServer := grpc.NewServer()
		linkRepository.RegisterLinkRepositoryServer(grpcServer, graph)
		_ = grpcServer.Serve(grpcListener)
	}()

	wg.Wait()
}

func getLinkGraph(linkGraphURI string) (linkRepository.LinkRepositoryServer, error) {
	if linkGraphURI == "" {
		return nil, errors.New("Link Graph URI not found")
	}
	return NewCockroachDbGraph(linkGraphURI)
}
