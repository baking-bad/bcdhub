package tests

import (
	"context"
	"time"
)

func (s *StorageTestSuite) TestBigMpActionsGet() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	actions, err := s.bigMapActions.Get(ctx, 4, 10, 0)
	s.Require().NoError(err)
	s.Require().Len(actions, 1)

	action := actions[0]
	s.Require().EqualValues(1, action.ID)
	s.Require().EqualValues(1, action.Action)
	s.Require().EqualValues(34, action.OperationID)
	s.Require().EqualValues(33, action.Level)
	s.Require().NotNil(action.SourcePtr)
	s.Require().EqualValues(4, *action.SourcePtr)
	s.Require().Nil(action.DestinationPtr)
	s.Require().EqualValues("KT1W3fGSo8XfRSESPAg3Jngzt3D8xpPqW64i", action.Address)
}
