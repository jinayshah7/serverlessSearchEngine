package dbspgraph

import (
	"errors"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/hashicorp/go-multierror"
	"github.com/jinayshah7/distributedSearchEngine/services/pagerank/dbspgraph/job"
)

type Serializer interface {
	Serialize(interface{}) (*any.Any, error)
	Unserialize(*any.Any) (interface{}, error)
}

type MasterConfig struct {
	ListenAddress string
	JobRunner     job.Runner
	Serializer    Serializer
}

func (cfg *MasterConfig) Validate() error {
	var err error
	if cfg.ListenAddress == "" {
		err = multierror.Append(err, errors.New("listen address not specified"))
	}
	if cfg.JobRunner == nil {
		err = multierror.Append(err, errors.New("job runner not specified"))
	}
	if cfg.Serializer == nil {
		err = multierror.Append(err, errors.New("aggregator serializer not specified"))
	}
	return err
}

type WorkerConfig struct {
	JobRunner  job.Runner
	Serializer Serializer
}

func (cfg *WorkerConfig) Validate() error {
	var err error
	if cfg.JobRunner == nil {
		err = multierror.Append(err, errors.New("job runner not specified"))
	}
	if cfg.Serializer == nil {
		err = multierror.Append(err, errors.New("message/aggregator serializer not specified"))
	}
	return err
}
