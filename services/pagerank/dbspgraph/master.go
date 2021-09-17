package dbspgraph

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/jinayshah7/distributedSearchEngine/services/pagerank/dbspgraph/job"
	"github.com/jinayshah7/distributedSearchEngine/services/pagerank/dbspgraph/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

var (
	minUUID = uuid.Nil
	maxUUID = uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")

	ErrUnableToReserveWorkers = errors.New("unable to reserve required number of workers")
)

type Master struct {
	cfg         MasterConfig
	workerPool  *workerPool
	srvListener net.Listener
}

func NewMaster(cfg MasterConfig) (*Master, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.New("master config validation failed: %w", err)
	}

	return &Master{
		cfg:        cfg,
		workerPool: newWorkerPool(),
	}, nil
}

func (m *Master) Start() error {
	var err error
	if m.srvListener, err = net.Listen("tcp", m.cfg.ListenAddress); err != nil {
		return errors.New("cannot start server: %w", err)
	}

	grpcService := grpc.NewServer()
	proto.RegisterJobQueueServer(grpcService, &masterRPCHandler{
		workerPool: m.workerPool,
	})
	log.Info("listening for worker connections")
	go func(l net.Listener) { _ = grpcService.Serve(l) }(m.srvListener)

	return nil
}

func (m *Master) Close() error {
	var err error

	if m.srvListener != nil {
		err = m.srvListener.Close()
		m.srvListener = nil
	}

	if cErr := m.workerPool.Close(); cErr != nil {
		err = multierror.Append(err, cErr)
	}

	return err
}

func (m *Master) RunJob(ctx context.Context, minWorkers int, workerAcquireTimeout time.Duration) error {
	var acquireCtx = ctx
	if workerAcquireTimeout != 0 {
		var cancelFn func()
		acquireCtx, cancelFn = context.WithTimeout(ctx, workerAcquireTimeout)
		defer cancelFn()
	}
	workers, err := m.workerPool.ReserveWorkers(acquireCtx, minWorkers)
	if err != nil {
		return ErrUnableToReserveWorkers
	}

	jobID := uuid.New().String()
	createdAt := time.Now().UTC().Truncate(time.Millisecond)
	coordinator, err := newMasterJobCoordinator(ctx, masterJobCoordinatorConfig{
		jobDetails: job.Details{
			JobID:           jobID,
			CreatedAt:       createdAt,
			PartitionFromID: minUUID,
			PartitionToID:   maxUUID,
		},
		workers:    workers,
		jobRunner:  m.cfg.JobRunner,
		serializer: m.cfg.Serializer,
	})
	if err != nil {
		err = errors.New("unable to create job coordinator: %w", err)
		for _, w := range workers {
			w.Close(err)
		}
		return err
	}

	log.Info("coordinating execution of new job")

	if err = coordinator.RunJob(); err != nil {
		log.Error("job execution failed %w", err)
		for _, w := range workers {
			w.Close(err)
		}
		return err
	}

	log.Info("job completed successfully")
	for _, w := range workers {
		w.Close(nil)
	}
	return nil
}

type masterRPCHandler struct {
	workerPool *workerPool
}

// JobStream implements JobQueueServer.
func (h *masterRPCHandler) JobStream(stream proto.JobQueue_JobStreamServer) error {
	workerStream := newRemoteWorkerStream(stream)
	h.workerPool.AddWorker(workerStream)
	return workerStream.HandleSendRecv()
}
