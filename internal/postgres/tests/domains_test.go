package tests

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/contract"
)

func (s *StorageTestSuite) TestSame() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	contracts, err := s.domains.Same(ctx, "public", contract.Contract{
		ID:        25,
		BabylonID: 5,
		AccountID: 132,
		Babylon: contract.Script{
			ID:   5,
			Hash: "621506fddbe82712919f68b8b52bcc684cff7bdc650409f4b038cffd2da1e018",
		},
	}, 2, 0, "public")
	s.Require().NoError(err)
	s.Require().Len(contracts, 2)
}

func (s *StorageTestSuite) TestSameCount() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	count, err := s.domains.SameCount(ctx, contract.Contract{
		ID:        25,
		BabylonID: 5,
		AccountID: 132,
		Babylon: contract.Script{
			ID:   5,
			Hash: "621506fddbe82712919f68b8b52bcc684cff7bdc650409f4b038cffd2da1e018",
		},
	}, "public")
	s.Require().NoError(err)
	s.Require().EqualValues(11, count)
}
