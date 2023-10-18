package tests

import (
	"context"
	"time"
)

func (s *StorageTestSuite) TestTableExists() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	exists := s.storage.TablesExist(ctx)
	s.Require().True(exists)
}
