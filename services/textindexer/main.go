package main

import (
	"context"
	"fmt"
	"github.com/jinayshah7/distributedSearchEngine/services/textindexer/es"
	"github.com/jinayshah7/distributedSearchEngine/services/textindexer/index"
	"github.com/jinayshah7/distributedSearchEngine/services/textindexer/textindexerapi"
	"github.com/jinayshah7/distributedSearchEngine/services/textindexer/textindexerapi/proto"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"strings"
	"sync"
)

var (
	appName = "services-textindexer"
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
			Name:   "text-indexer-uri",
			Value:  "",
			EnvVar: "TEXT_INDEXER_URI",
			Usage:  "The URI for connecting to the text indexer",
		},
		cli.IntFlag{
			Name:   "grpc-port",
			Value:  8080,
			EnvVar: "GRPC_PORT",
			Usage:  "The port for exposing the gRPC endpoints for accessing the text indexer",
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

	indexer, err := getTextIndexer(appCtx.String("text-indexer-uri"))
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
		proto.RegisterTextIndexerServer(srv, textindexerapi.NewTextIndexerServer(indexer))
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

func getTextIndexer(textIndexerURI string) (index.Indexer, error) {
	if textIndexerURI == "" {
		return nil, xerrors.Errorf("text indexer URI must be specified with --text-indexer-uri")
	}

	uri, err := url.Parse(textIndexerURI)
	if err != nil {
		return nil, xerrors.Errorf("could not parse text indexer URI: %w", err)
	}

	nodes := strings.Split(uri.Host, ",")
	for i := 0; i < len(nodes); i++ {
		nodes[i] = "http://" + nodes[i]
	}
	logger.Info("using ES indexer")
	return es.NewElasticSearchIndexer(nodes, false)

}
