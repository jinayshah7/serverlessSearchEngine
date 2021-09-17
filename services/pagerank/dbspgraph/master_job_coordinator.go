package dbspgraph

import (
	"context"
	"sync"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"github.com/jinayshah7/distributedSearchEngine/services/pagerank/bspgraph"
	"github.com/jinayshah7/distributedSearchEngine/services/pagerank/dbspgraph/job"
	"github.com/jinayshah7/distributedSearchEngine/services/pagerank/dbspgraph/partition"
	"github.com/jinayshah7/distributedSearchEngine/services/pagerank/dbspgraph/proto"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

// masterJobCoordinatorConfig encapsulates the configuration options for the
// master job coordinator.
type masterJobCoordinatorConfig struct {
	jobDetails job.Details
	workers    []*remoteWorkerStream

	jobRunner  job.Runner
	serializer Serializer
	logger     *logrus.Entry
}

// masterJobCoordinator is used by the master node to coordinate the individual
// worker instances so the various job stages can be executed in lock-step.
type masterJobCoordinator struct {
	jobCtx       context.Context
	cancelJobCtx func()

	barrier   *masterStepBarrier
	partRange *partition.Range

	cfg masterJobCoordinatorConfig
}

func newMasterJobCoordinator(ctx context.Context, cfg masterJobCoordinatorConfig) (*masterJobCoordinator, error) {
	partRange, err := partition.NewRange(cfg.jobDetails.PartitionFromID, cfg.jobDetails.PartitionToID, len(cfg.workers))
	if err != nil {
		return nil, err
	}

	jobCtx, cancelJobCtx := context.WithCancel(ctx)
	return &masterJobCoordinator{
		jobCtx:       jobCtx,
		cancelJobCtx: cancelJobCtx,
		barrier:      newMasterStepBarrier(jobCtx, len(cfg.workers)),
		partRange:    partRange,
		cfg:          cfg,
	}, nil
}

func (c *masterJobCoordinator) RunJob() error {
	// Create a wrapper for patching the user-defined executor callbacks so
	// they can be executed in lock-step with the workers and pass the
	// resulting factory to the job runner to get back an Executor for the
	// graph.
	execFactory := newMasterExecutorFactory(c.cfg.serializer, c.barrier)
	executor, err := c.cfg.jobRunner.StartJob(c.cfg.jobDetails, execFactory)
	if err != nil {
		c.cancelJobCtx()
		return xerrors.Errorf("unable to start job on master: %w", err)
	}

	for assignedPartition, w := range c.cfg.workers {
		w.SetDisconnectCallback(c.handleWorkerDisconnect)
		if err := c.publishJobDetails(w, assignedPartition); err != nil {
			c.cfg.jobRunner.AbortJob(c.cfg.jobDetails)
			c.cancelJobCtx()
			return err
		}
	}

	// Start a goroutine to process incoming messages from each worker.
	var wg sync.WaitGroup
	wg.Add(len(c.cfg.workers))
	graph := executor.Graph()
	for workerIndex, worker := range c.cfg.workers {
		go func(workerIndex int, worker *remoteWorkerStream) {
			defer wg.Done()
			c.handleWorkerPayloads(workerIndex, worker, graph)
		}(workerIndex, worker)
	}

	if err = c.runJobToCompletion(executor); err != nil {
		c.cfg.jobRunner.AbortJob(c.cfg.jobDetails)
		if xerrors.Is(err, context.Canceled) {
			err = errJobAborted
		}
	}

	c.cancelJobCtx()
	wg.Wait() // wait for any spawned goroutines to exit before returning.
	return err
}

// handleWorkerDisconnect is invoked when a remote worker stream disconnects.
func (c *masterJobCoordinator) handleWorkerDisconnect() {
	select {
	case <-c.jobCtx.Done(): // job already aborted
	default:
		c.cfg.logger.Error("lost connection to worker; aborting job")
		c.cancelJobCtx()
	}
}

// publishJobDetails figures out the UUID range assignment for a remote worker
// and writes a JobDetails message to its stream.
func (c *masterJobCoordinator) publishJobDetails(w *remoteWorkerStream, assignedPartition int) error {
	partitionFromID, partitionToID, err := c.partRange.PartitionExtents(assignedPartition)
	if err != nil {
		return xerrors.Errorf("unable to calculate partition assignment: %w", err)
	}

	ts, err := ptypes.TimestampProto(c.cfg.jobDetails.CreatedAt)
	if err != nil {
		return xerrors.Errorf("unable to encode job creation time: %w", err)
	}

	c.sendToWorker(w, &proto.MasterPayload{
		Payload: &proto.MasterPayload_JobDetails{
			JobDetails: &proto.JobDetails{
				JobId:             c.cfg.jobDetails.JobID,
				CreatedAt:         ts,
				PartitionFromUuid: partitionFromID[:],
				PartitionToUuid:   partitionToID[:],
			},
		},
	})
	return nil
}

func (c *masterJobCoordinator) runJobToCompletion(executor *bspgraph.Executor) error {
	if err := executor.RunToCompletion(c.jobCtx); err != nil {
		return err
	} else if _, err := c.barrier.WaitForWorkers(proto.Step_EXECUTED_GRAPH); err != nil {
		return err
	} else if err := c.barrier.NotifyWorkers(&proto.Step{Type: proto.Step_EXECUTED_GRAPH}); err != nil {
		return err
	} else if err := c.cfg.jobRunner.CompleteJob(c.cfg.jobDetails); err != nil {
		return err
	} else if _, err := c.barrier.WaitForWorkers(proto.Step_PESISTED_RESULTS); err != nil {
		return err
	} else if err := c.barrier.NotifyWorkers(&proto.Step{Type: proto.Step_PESISTED_RESULTS}); err != nil {
		return err
	} else if _, err := c.barrier.WaitForWorkers(proto.Step_COMPLETED_JOB); err != nil {
		return err
	}

	return nil
}

func (c *masterJobCoordinator) handleWorkerPayloads(workerIndex int, worker *remoteWorkerStream, graph *bspgraph.Graph) {
	var wPayload *proto.WorkerPayload
	for {
		select {
		case wPayload = <-worker.RecvFromWorkerChan():
		case <-c.jobCtx.Done():
			return
		}

		if relayMsg := wPayload.GetRelayMessage(); relayMsg != nil {
			c.relayMessageToWorker(workerIndex, relayMsg)
		} else if stepMsg := wPayload.GetStep(); stepMsg != nil {
			// Enter the barrier and wait for master's notification.
			updatedStep, err := c.barrier.Wait(stepMsg)
			if err != nil {
				c.cancelJobCtx()
				return
			}

			// Send updated step back to the worker.
			c.sendToWorker(worker, &proto.MasterPayload{
				Payload: &proto.MasterPayload_Step{Step: updatedStep},
			})
		}
	}
}

func (c *masterJobCoordinator) relayMessageToWorker(srcWorkerIndex int, relayMsg *proto.RelayMessage) {
	dstUUID, err := uuid.Parse(relayMsg.Destination)
	if err != nil {
		log.Error("unable to parse message destination UUID %w", err)
		c.cancelJobCtx()
		return
	}

	partIndex, err := c.partRange.PartitionForID(dstUUID)
	if err != nil {
		log.Error("unable to identify target partition for message %w", err)
		c.cancelJobCtx()
		return
	}

	// If the message destination is the same worker that asked us to relay
	// it in the first place, assume that the destination is invalid.
	if partIndex == srcWorkerIndex {
		c.cfg.logger.WithField("dst_id", relayMsg.Destination).Error("received relay request for message to a vertex that does not exist")
		c.cancelJobCtx()
		return
	}

	c.sendToWorker(c.cfg.workers[partIndex], &proto.MasterPayload{
		Payload: &proto.MasterPayload_RelayMessage{RelayMessage: relayMsg},
	})
}

func (c *masterJobCoordinator) sendToWorker(worker *remoteWorkerStream, mMsg *proto.MasterPayload) {
	select {
	case worker.SendToWorkerChan() <- mMsg:
	case <-c.jobCtx.Done():
	}
}
