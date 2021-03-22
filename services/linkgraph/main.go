package main

import (
	"context"
	"fmt"
	"github.com/jinayshah7/distributedSearchEngine/services/linkgraph/cdb"
	"github.com/jinayshah7/distributedSearchEngine/services/linkgraph/graph"
	"github.com/jinayshah7/distributedSearchEngine/services/linkgraph/linkgraphapi"
	"github.com/jinayshah7/distributedSearchEngine/services/linkgraph/linkgraphapi/proto"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"
)

var (
	appName = "services-linkgraph"
	logger  *logrus.Entry
)

func main() {
	host, _ := os.Hostname()
	rootLogger := logrus.New()
	rootLogger.SetFormatter(new(logrus.JSONFormatter))
	logger = rootLogger.WithFields(logrus.Fields{
		"app":  appName,
		"host": host,
	})

	if err := makeApp().Run(os.Args); err != nil {
		logger.WithField("err", err).Error("shutting down due to error")
		_ = os.Stderr.Sync()
		os.Exit(1)
	}
}

func makeApp() *cli.App {
	app := cli.NewApp()
	app.Name = appName
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "link-graph-uri",
			Value:  "",
			EnvVar: "LINK_GRAPH_URI",
			Usage:  "The URI for connecting to the link-graph",
		},
		cli.IntFlag{
			Name:   "grpc-port",
			Value:  8080,
			EnvVar: "GRPC_PORT",
			Usage:  "The port for exposing the gRPC endpoints for accessing the link graph",
		},
		cli.IntFlag{
			Name:   "pprof-port",
			Value:  6060,
			EnvVar: "PPROF_PORT",
			Usage:  "The port for exposing pprof endpoints",
		},
	}
	app.Action = runMain
	return app
}

func runMain(appCtx *cli.Context) error {
	var wg sync.WaitGroup
	_, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	graph, err := getLinkGraph(appCtx.String("link-graph-uri"))
	if err != nil {
		return err
	}

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", appCtx.Int("grpc-port")))
	if err != nil {
		return err
	}
	defer func() { _ = grpcListener.Close() }()

	wg.Add(1)
	go func() {
		defer wg.Done()
		srv := grpc.NewServer()
		proto.RegisterLinkGraphServer(srv, linkgraphapi.NewLinkGraphServer(graph))
		logger.WithField("port", appCtx.Int("grpc-port")).Info("listening for gRPC connections")
		_ = srv.Serve(grpcListener)
	}()

	pprofListener, err := net.Listen("tcp", fmt.Sprintf(":%d", appCtx.Int("pprof-port")))
	if err != nil {
		return err
	}
	defer func() { _ = pprofListener.Close() }()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.WithField("port", appCtx.Int("pprof-port")).Info("listening for pprof requests")
		srv := new(http.Server)
		_ = srv.Serve(pprofListener)
	}()

	wg.Wait()
	return nil
}

func getLinkGraph(linkGraphURI string) (graph.Graph, error) {
	if linkGraphURI == "" {
		return nil, xerrors.Errorf("link graph URI must be specified with --link-graph-uri")
	}
	return cdb.NewCockroachDbGraph(linkGraphURI)
}
