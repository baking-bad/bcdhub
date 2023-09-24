package tests

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
)

func (s *StorageTestSuite) TestMigrationGet() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	migrations, err := s.migrations.Get(ctx, 1)
	s.Require().NoError(err)
	s.Require().Len(migrations, 1)

	m := migrations[0]
	s.Require().EqualValues(1, m.ID)
	s.Require().EqualValues(2, m.ProtocolID)
	s.Require().EqualValues(0, m.PrevProtocolID)
	s.Require().EqualValues(1, m.ContractID)
	s.Require().EqualValues(2, m.Level)
	s.Require().EqualValues(types.MigrationKindBootstrap, m.Kind)
}
