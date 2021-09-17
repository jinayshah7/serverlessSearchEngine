package dbspgraph

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"github.com/jinayshah7/distributedSearchEngine/services/pagerank/dbspgraph/job"
	"github.com/jinayshah7/distributedSearchEngine/services/pagerank/dbspgraph/proto"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
)

type Worker struct {
	cfg WorkerConfig

	masterConn *grpc.ClientConn
	masterCli  proto.JobQueueClient
}

// NewWorker creates a new Worker instance with the specified configuration.
func NewWorker(cfg WorkerConfig) (*Worker, error) {
	if err := cfg.Validate(); err != nil {
		return nil, xerrors.Errorf("worker config validation failed: %w", err)
	}

	return &Worker{cfg: cfg}, nil
}

// Dial establishes a connection to the master node.
func (w *Worker) Dial(masterEndpoint string, dialTimeout time.Duration) error {
	var dialCtx context.Context
	if dialTimeout != 0 {
		var cancelFn func()
		dialCtx, cancelFn = context.WithTimeout(context.Background(), dialTimeout)
		defer cancelFn()
	}

	conn, err := grpc.DialContext(dialCtx, masterEndpoint, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return xerrors.Errorf("unable to dial master: %w", err)
	}

	w.masterConn = conn
	w.masterCli = proto.NewJobQueueClient(conn)
	return nil
}

func (w *Worker) Close() error {
	var err error
	if w.masterConn != nil {
		err = w.masterConn.Close()
		w.masterConn = nil
	}
	return err
}

func (w *Worker) RunJob(ctx context.Context) error {
	stream, err := w.masterCli.JobStream(ctx)
	if err != nil {
		return err
	}

	log.Info("waiting for next job")
	jobDetails, err := w.waitForJob(stream)
	if err != nil {
		return err
	}

	masterStream := newRemoteMasterStream(stream)
	jobLogger := w.cfg.Logger.WithField("job_id", jobDetails.JobID)
	coordinator := newWorkerJobCoordinator(ctx, workerJobCoordinatorConfig{
		jobDetails:   jobDetails,
		masterStream: masterStream,
		jobRunner:    w.cfg.JobRunner,
		serializer:   w.cfg.Serializer,
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := masterStream.HandleSendRecv(); err != nil {
			coordinator.cancelJobCtx()
		}
	}()
	log.Info("starting new job")

	if err = coordinator.RunJob(); err != nil {
		log.Error("job execution failed %w", err)
	} else {
		log.Info("job completed successfully")
	}
	masterStream.Close()
	wg.Wait()
	return err
}

func (w *Worker) waitForJob(jobStream proto.JobQueue_JobStreamClient) (job.Details, error) {
	var jobDetails job.Details

	mMsg, err := jobStream.Recv()
	if err != nil {
		return jobDetails, errors.New("unable to read job details from master: %w", err)
	}

	jobDetailsMsg := mMsg.GetJobDetails()
	if jobDetailsMsg == nil {
		return jobDetails, errors.New("expected master to send a JobDetails message")
	}

	jobDetails.JobID = jobDetailsMsg.JobId
	if jobDetails.CreatedAt, err = ptypes.Timestamp(jobDetailsMsg.CreatedAt); err != nil {
		return jobDetails, errors.New("unable to parse job creation time: %w", err)
	} else if jobDetails.PartitionFromID, err = uuid.FromBytes(jobDetailsMsg.PartitionFromUuid[:]); err != nil {
		return jobDetails, errors.New("unable to parse partition start UUID: %w", err)
	} else if jobDetails.PartitionToID, err = uuid.FromBytes(jobDetailsMsg.PartitionToUuid[:]); err != nil {
		return jobDetails, errors.New("unable to parse partition end UUID: %w", err)
	}

	return jobDetails, nil
}
