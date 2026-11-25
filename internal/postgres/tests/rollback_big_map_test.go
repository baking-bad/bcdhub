package tests

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/baking-bad/bcdhub/internal/rollback"
)

// The big map rollback tests run the real rollback.Manager against Postgres.
// Each one lays out the history of a dedicated pointer below rollbackLevel,
// writes block rollbackLevel the way the indexer would, rolls it back and
// checks big_map_states.

const (
	rollbackLevel    = int64(1000)
	rollbackContract = "KT1RollbackBigMapTest1111111111111111"
)

func levelTime(level int64) time.Time {
	return time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(level) * time.Minute)
}

type bigMapHistory struct {
	s   *StorageTestSuite
	ptr int64
}

func (s *StorageTestSuite) newBigMapHistory(ctx context.Context, ptr int64) bigMapHistory {
	_, err := s.storage.DB.NewInsert().Model(&block.Block{
		Hash:       "BLRollbackBigMapTest",
		Level:      rollbackLevel,
		Timestamp:  levelTime(rollbackLevel),
		ProtocolID: 3,
	}).Exec(ctx)
	s.Require().NoError(err)
	return bigMapHistory{s: s, ptr: ptr}
}

// diff writes a big map diff; nil value means the key was removed.
func (h bigMapHistory) diff(ctx context.Context, keyHash string, level int64, value types.Bytes) {
	_, err := h.s.storage.DB.NewInsert().Model(&bigmapdiff.BigMapDiff{
		Ptr:         h.ptr,
		Key:         types.Bytes(`{"string":"` + keyHash + `"}`),
		KeyHash:     keyHash,
		Value:       value,
		Level:       level,
		Contract:    rollbackContract,
		Timestamp:   levelTime(level),
		ProtocolID:  3,
		OperationID: level,
	}).Exec(ctx)
	h.s.Require().NoError(err)
}

// state writes the current state of a key as it is after block rollbackLevel.
func (h bigMapHistory) state(ctx context.Context, keyHash string, level int64, value types.Bytes, removed bool) {
	_, err := h.s.storage.DB.NewInsert().Model(&bigmapdiff.BigMapState{
		Ptr:             h.ptr,
		Key:             types.Bytes(`{"string":"` + keyHash + `"}`),
		KeyHash:         keyHash,
		Value:           value,
		Removed:         removed,
		Contract:        rollbackContract,
		LastUpdateLevel: level,
		LastUpdateTime:  levelTime(level),
	}).Exec(ctx)
	h.s.Require().NoError(err)
}

func (h bigMapHistory) rollback(ctx context.Context) {
	saver, err := postgres.NewRollback(h.s.storage.DB)
	h.s.Require().NoError(err)

	manager := rollback.NewManager(h.s.storage, h.s.blocks, saver, h.s.stats)
	err = manager.Rollback(ctx, types.Mainnet, block.Block{Level: rollbackLevel}, rollbackLevel-1)
	h.s.Require().NoError(err)

	diffsLeft, err := h.s.storage.DB.NewSelect().
		Model((*bigmapdiff.BigMapDiff)(nil)).
		Where("ptr = ?", h.ptr).
		Where("level >= ?", rollbackLevel).
		Count(ctx)
	h.s.Require().NoError(err)
	h.s.Require().Zero(diffsLeft, "diffs of the rolled back block must be deleted")
}

func (h bigMapHistory) states(ctx context.Context) map[string]bigmapdiff.BigMapState {
	var states []bigmapdiff.BigMapState
	err := h.s.storage.DB.NewSelect().Model(&states).Where("ptr = ?", h.ptr).Scan(ctx)
	h.s.Require().NoError(err)

	result := make(map[string]bigmapdiff.BigMapState, len(states))
	for i := range states {
		result[states[i].KeyHash] = states[i]
	}
	return result
}

func (h bigMapHistory) requireState(states map[string]bigmapdiff.BigMapState, keyHash string, level int64, value string, removed bool) {
	state, ok := states[keyHash]
	h.s.Require().True(ok, "state %s must exist", keyHash)
	h.s.Require().Equal(value, string(state.Value), keyHash)
	h.s.Require().Equal(removed, state.Removed, keyHash)
	h.s.Require().EqualValues(level, state.LastUpdateLevel, keyHash)
	h.s.Require().True(levelTime(level).Equal(state.LastUpdateTime), "%s: last_update_time %s", keyHash, state.LastUpdateTime)
}

func (s *StorageTestSuite) TestRollbackBigMap_UpdatedKeyGetsPreviousValue() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	h := s.newBigMapHistory(ctx, 10001)
	h.diff(ctx, "key", 900, types.Bytes(`{"int":"1"}`))
	h.diff(ctx, "key", 950, types.Bytes(`{"int":"2"}`))
	h.diff(ctx, "key", rollbackLevel, types.Bytes(`{"int":"3"}`))
	h.state(ctx, "key", rollbackLevel, types.Bytes(`{"int":"3"}`), false)

	h.rollback(ctx)

	states := h.states(ctx)
	s.Require().Len(states, 1)
	h.requireState(states, "key", 950, `{"int":"2"}`, false)
}

func (s *StorageTestSuite) TestRollbackBigMap_KeyCreatedInBlockIsDeleted() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	h := s.newBigMapHistory(ctx, 10002)
	h.diff(ctx, "key", rollbackLevel, types.Bytes(`{"int":"1"}`))
	h.state(ctx, "key", rollbackLevel, types.Bytes(`{"int":"1"}`), false)

	h.rollback(ctx)

	s.Require().Empty(h.states(ctx))
}

func (s *StorageTestSuite) TestRollbackBigMap_RemovedKeyIsRestored() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	h := s.newBigMapHistory(ctx, 10003)
	h.diff(ctx, "key", 900, types.Bytes(`{"int":"1"}`))
	h.diff(ctx, "key", rollbackLevel, nil)
	// The upsert keeps the previous value of a removed key.
	h.state(ctx, "key", rollbackLevel, types.Bytes(`{"int":"1"}`), true)

	h.rollback(ctx)

	states := h.states(ctx)
	s.Require().Len(states, 1)
	h.requireState(states, "key", 900, `{"int":"1"}`, false)
}

func (s *StorageTestSuite) TestRollbackBigMap_WholeBigMapRemovalIsRestored() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	h := s.newBigMapHistory(ctx, 10004)
	h.diff(ctx, "alive_1", 900, types.Bytes(`{"int":"1"}`))
	h.diff(ctx, "alive_2", 850, types.Bytes(`{"int":"2"}`))
	h.diff(ctx, "alive_2", 950, types.Bytes(`{"int":"22"}`))
	// This key had already been removed before the big map was.
	h.diff(ctx, "removed_before", 800, types.Bytes(`{"int":"3"}`))
	h.diff(ctx, "removed_before", 900, nil)

	// Removing the whole big map at rollbackLevel writes a null diff for every
	// key and stamps every state with the removal level.
	for _, keyHash := range []string{"alive_1", "alive_2", "removed_before"} {
		h.diff(ctx, keyHash, rollbackLevel, nil)
	}
	h.state(ctx, "alive_1", rollbackLevel, types.Bytes(`{"int":"1"}`), true)
	h.state(ctx, "alive_2", rollbackLevel, types.Bytes(`{"int":"22"}`), true)
	h.state(ctx, "removed_before", rollbackLevel, types.Bytes(`{"int":"3"}`), true)

	h.rollback(ctx)

	states := h.states(ctx)
	s.Require().Len(states, 3)
	h.requireState(states, "alive_1", 900, `{"int":"1"}`, false)
	h.requireState(states, "alive_2", 950, `{"int":"22"}`, false)
	h.requireState(states, "removed_before", 900, `{"int":"3"}`, true)
}

// Same as above, but the removal goes through the indexer's write path:
// states are read with GetByPtr and written back with the
// Transaction.BigMapStates upsert, stamped the way the removal handlers in
// internal/parsers/storage do it.
func (s *StorageTestSuite) TestRollbackBigMap_WholeBigMapRemovalThroughUpsert() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	h := s.newBigMapHistory(ctx, 10006)
	h.diff(ctx, "alive", 900, types.Bytes(`{"int":"1"}`))
	h.state(ctx, "alive", 900, types.Bytes(`{"int":"1"}`), false)
	h.diff(ctx, "removed_before", 800, types.Bytes(`{"int":"2"}`))
	h.diff(ctx, "removed_before", 850, nil)
	h.state(ctx, "removed_before", 850, types.Bytes(`{"int":"2"}`), true)

	states, err := s.bigMapDiffs.GetByPtr(ctx, rollbackContract, h.ptr)
	s.Require().NoError(err)
	s.Require().Len(states, 2)

	removed := make([]*bigmapdiff.BigMapState, len(states))
	for i := range states {
		states[i].Removed = true
		states[i].LastUpdateLevel = rollbackLevel
		states[i].LastUpdateTime = levelTime(rollbackLevel)
		removed[i] = &states[i]
		h.diff(ctx, states[i].KeyHash, rollbackLevel, nil)
	}

	tx, err := core.NewTransaction(ctx, s.storage.DB)
	s.Require().NoError(err)
	defer tx.Close()
	s.Require().NoError(tx.BigMapStates(ctx, removed...))
	s.Require().NoError(tx.Commit())

	afterRemoval := h.states(ctx)
	h.requireState(afterRemoval, "alive", rollbackLevel, `{"int":"1"}`, true)
	h.requireState(afterRemoval, "removed_before", rollbackLevel, `{"int":"2"}`, true)

	h.rollback(ctx)

	restored := h.states(ctx)
	s.Require().Len(restored, 2)
	h.requireState(restored, "alive", 900, `{"int":"1"}`, false)
	h.requireState(restored, "removed_before", 850, `{"int":"2"}`, true)
}

func (s *StorageTestSuite) TestLastDiffIgnoresRolledBackLevel() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	h := s.newBigMapHistory(ctx, 10005)
	h.diff(ctx, "key", 900, types.Bytes(`{"int":"1"}`))
	h.diff(ctx, "key", 950, nil)
	h.diff(ctx, "key", rollbackLevel, types.Bytes(`{"int":"2"}`))

	saver, err := postgres.NewRollback(s.storage.DB)
	s.Require().NoError(err)
	defer func() { s.Require().NoError(saver.Close()) }()

	diff, err := saver.LastDiff(ctx, h.ptr, "key", rollbackLevel, false)
	s.Require().NoError(err)
	s.Require().EqualValues(950, diff.Level)
	s.Require().Nil(diff.Value)

	diff, err = saver.LastDiff(ctx, h.ptr, "key", rollbackLevel, true)
	s.Require().NoError(err)
	s.Require().EqualValues(900, diff.Level)
	s.Require().Equal(`{"int":"1"}`, string(diff.Value))

	_, err = saver.LastDiff(ctx, h.ptr, "key", 900, false)
	s.Require().True(s.storage.IsRecordNotFound(err), "no diff below the first one: %v", err)
}
