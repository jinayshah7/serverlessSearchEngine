package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
  "strconv"

	"github.com/jinayshah7/distributedSearchEngine/proto/linkRepository"
	"google.golang.org/grpc"
)

func main() {
	var wg sync.WaitGroup
	grpcPort, _ := strconv.Atoi(os.Getenv("GRPC_PORT"))
	linkRepositoryURL := os.Getenv("LINK_REPOSITORY_URL")

	repository, err := getLinkRepository(linkRepositoryURL)
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
		linkRepository.RegisterLinkRepositoryServer(grpcServer, repository)
		_ = grpcServer.Serve(grpcListener)
	}()

	wg.Wait()
}

func getLinkRepository(linkRepositoryURL string) (linkRepository.LinkRepositoryServer, error) {
	if linkRepositoryURL == "" {
		return nil, errors.New("Link repository URL not found")
	}
	return NewCockroachDbRepository(linkRepositoryURL)
}
