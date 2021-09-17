package partition

import (
	"bytes"
	"math/big"
	"sort"

	"github.com/google/uuid"
)

type Range struct {
	start       uuid.UUID
	rangeSplits []uuid.UUID
}

func NewFullRange(numPartitions int) (*Range, error) {
	return NewRange(
		uuid.Nil,
		uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"),
		numPartitions,
	)
}

func NewRange(start, end uuid.UUID, numPartitions int) (*Range, error) {
	if bytes.Compare(start[:], end[:]) >= 0 {
		return nil, errors.New("range start UUID must be less than the end UUID")
	} else if numPartitions <= 0 {
		return nil, errors.New("number of partitions must be at least equal to 1")
	}

	tokenRange := big.NewInt(0)
	partSize := big.NewInt(0)
	partSize = partSize.Sub(big.NewInt(0).SetBytes(end[:]), big.NewInt(0).SetBytes(start[:]))
	partSize = partSize.Div(partSize.Add(partSize, big.NewInt(1)), big.NewInt(int64(numPartitions)))

	var (
		to     uuid.UUID
		err    error
		ranges = make([]uuid.UUID, numPartitions)
	)
	for partition := 0; partition < numPartitions; partition++ {
		if partition == numPartitions-1 {
			to = end
		} else {
			tokenRange.Mul(partSize, big.NewInt(int64(partition+1)))
			if to, err = uuid.FromBytes(tokenRange.Bytes()); err != nil {
				return nil, xerrors.Errorf("partition range: %w", err)
			}
		}

		ranges[partition] = to
	}

	return &Range{start: start, rangeSplits: ranges}, nil
}

func (r *Range) Extents() (uuid.UUID, uuid.UUID) {
	return r.start, r.rangeSplits[len(r.rangeSplits)-1]
}

func (r *Range) PartitionExtents(partition int) (uuid.UUID, uuid.UUID, error) {
	if partition < 0 || partition >= len(r.rangeSplits) {
		return uuid.Nil, uuid.Nil, xerrors.Errorf("invalid partition index")
	}

	if partition == 0 {
		return r.start, r.rangeSplits[0], nil
	}
	return r.rangeSplits[partition-1], r.rangeSplits[partition], nil
}

func (r *Range) PartitionForID(id uuid.UUID) (int, error) {
	partIndex := sort.Search(len(r.rangeSplits), func(n int) bool {
		return bytes.Compare(id[:], r.rangeSplits[n][:]) < 0
	})

	if bytes.Compare(id[:], r.start[:]) < 0 || partIndex >= len(r.rangeSplits) {
		return -1, errors.New("unable to detect partition for ID %q", id)
	}

	return partIndex, nil
}
