package main

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/jinayshah7/distributedSearchEngine/services/linkgraph/cdb"
	"github.com/jinayshah7/distributedSearchEngine/services/linkgraph/graph"
	"github.com/jinayshah7/distributedSearchEngine/services/linkgraph/linkgraphapi"
	"github.com/jinayshah7/distributedSearchEngine/services/linkgraph/linkgraphapi/proto"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

func main() {
}

func runMain(appCtx *cli.Context) error {
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
		srv := grpc.NewServer()
		proto.RegisterLinkGraphServer(srv, linkgraphapi.NewLinkGraphServer(graph))
		_ = srv.Serve(grpcListener)
	}()

	wg.Wait()
	return nil
}

func getLinkGraph(linkGraphURI string) (graph.Graph, error) {
	if linkGraphURI == "" {
		return nil, errors.New("Link Graph URI not found")
	}
	return cdb.NewCockroachDbGraph(linkGraphURI)
}
