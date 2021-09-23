package main

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/jinayshah7/distributedSearchEngine/services/linkrepository/cdb"
	"google.golang.org/grpc"
)

func main() {
	var wg sync.WaitGroup
	grpcPort := os.GetEnv("GRPC_PORT")
	linkGraphURI := os.GetEnv("LINK_GRAPH_URI")

	graph, err := getLinkGraph(linkGraphURI)
	if err != nil {
		return err
	}

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		return err
	}
	defer func() { _ = grpcListener.Close() }()

	wg.Add(1)
	go func() {
		defer wg.Done()
		grpcServer := grpc.NewServer()
		proto.RegisterLinkGraphServer(srv, graph)
		_ = grpcServer.Serve(grpcListener)
	}()

	wg.Wait()
}

func getLinkGraph(linkGraphURI string) (linkrepository.linkGraphClient, error) {
	if linkGraphURI == "" {
		return nil, errors.New("Link Graph URI not found")
	}
	return cdb.NewCockroachDbGraph(linkGraphURI)
}
