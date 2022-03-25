package operations

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	astContract "github.com/baking-bad/bcdhub/internal/bcd/contract"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func newInt64Ptr(val int64) *int64 {
	return &val
}

func readJSONFile(name string, response interface{}) error {
	bytes, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, response)
}

func readTestScript(address, symLink string) ([]byte, error) {
	path := filepath.Join("./test/contracts", fmt.Sprintf("%s_%s.json", address, symLink))
	return ioutil.ReadFile(path)
}

func readRPCScript(_ context.Context, address string, _ int64) (noderpc.Script, error) {
	var script noderpc.Script
	storageFile := fmt.Sprintf("./data/rpc/script/script/%s.json", address)
	if _, err := os.Lstat(storageFile); !os.IsNotExist(err) {
		f, err := os.Open(storageFile)
		if err != nil {
			return script, err
		}
		defer f.Close()

		err = json.NewDecoder(f).Decode(&script)
		return script, err
	}
	return script, errors.Errorf("unknown RPC script: %s", address)
}

func readTestScriptModel(address, symLink string) (contract.Script, error) {
	data, err := readTestScript(address, bcd.SymLinkBabylon)
	if err != nil {
		return contract.Script{}, err
	}
	var buffer bytes.Buffer
	buffer.WriteString(`{"code":`)
	buffer.Write(data)
	buffer.WriteString(`,"storage":{}}`)
	script, err := astContract.NewParser(buffer.Bytes())
	if err != nil {
		return contract.Script{}, errors.Wrap(err, "astContract.NewParser")
	}
	if err := script.Parse(); err != nil {
		return contract.Script{}, err
	}
	var s bcd.RawScript
	if err := json.Unmarshal(data, &s); err != nil {
		return contract.Script{}, err
	}
	return contract.Script{
		Code:        s.Code,
		Parameter:   s.Parameter,
		Storage:     s.Storage,
		Hash:        script.Hash,
		FailStrings: script.FailStrings.Values(),
		Annotations: script.Annotations.Values(),
		Tags:        types.NewTags(script.Tags.Values()),
		Hardcoded:   script.HardcodedAddresses.Values(),
	}, nil
}

//nolint
func readTestScriptPart(address, symLink, part string) ([]byte, error) {
	data, err := readTestScript(address, bcd.SymLinkBabylon)
	if err != nil {
		return nil, err
	}
	var s bcd.RawScript
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	switch part {
	case consts.CODE:
		return s.Code, nil
	case consts.PARAMETER:
		return s.Parameter, nil
	case consts.STORAGE:
		return s.Storage, nil
	}
	return nil, nil
}

func readTestContractModel(address string) (contract.Contract, error) {
	var c contract.Contract
	f, err := os.Open(fmt.Sprintf("./data/models/contract/%s.json", address))
	if err != nil {
		return c, err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&c)
	return c, err
}

func readStorage(_ context.Context, address string, level int64) ([]byte, error) {
	storageFile := fmt.Sprintf("./data/rpc/script/storage/%s_%d.json", address, level)
	return ioutil.ReadFile(storageFile)
}

func compareParserResponse(t *testing.T, got, want *parsers.TestStore) bool {
	if !assert.Len(t, got.BigMapState, len(want.BigMapState)) {
		return false
	}
	if !assert.Len(t, got.Contracts, len(want.Contracts)) {
		return false
	}
	if !assert.Len(t, got.Migrations, len(want.Migrations)) {
		return false
	}
	if !assert.Len(t, got.Operations, len(want.Operations)) {
		return false
	}
	if !assert.Len(t, got.TokenBalances, len(want.TokenBalances)) {
		return false
	}
	if !assert.Len(t, got.GlobalConstants, len(want.GlobalConstants)) {
		return false
	}

	for i := range got.Contracts {
		if !compareContract(t, want.Contracts[i], got.Contracts[i]) {
			return false
		}
	}
	for i := range got.Migrations {
		if !assert.Equal(t, want.Migrations[i], got.Migrations[i]) {
			return false
		}
	}
	for i := range got.Operations {
		if !compareOperations(t, want.Operations[i], got.Operations[i]) {
			return false
		}
	}
	for i := range got.TokenBalances {
		if !assert.Equal(t, want.TokenBalances[i], got.TokenBalances[i]) {
			return false
		}
	}
	for i := range got.BigMapState {
		if !assert.Equal(t, want.BigMapState[i], got.BigMapState[i]) {
			return false
		}
	}
	for i := range got.GlobalConstants {
		if !assert.Equal(t, want.GlobalConstants[i], got.GlobalConstants[i]) {
			return false
		}
	}

	return true
}

func compareTransfers(t *testing.T, one, two *transfer.Transfer) bool {
	if one.Contract != two.Contract {
		logger.Info().Msgf("Contract: %s != %s", one.Contract, two.Contract)
		return false
	}
	if !assert.Equal(t, one.Initiator, two.Initiator) {
		return false
	}
	if one.Status != two.Status {
		logger.Info().Msgf("Status: %s != %s", one.Status, two.Status)
		return false
	}
	if one.Timestamp != two.Timestamp {
		logger.Info().Msgf("Timestamp: %s != %s", one.Timestamp, two.Timestamp)
		return false
	}
	if one.Level != two.Level {
		logger.Info().Msgf("Level: %d != %d", one.Level, two.Level)
		return false
	}
	if !assert.Equal(t, one.From, two.From) {
		return false
	}
	if !assert.Equal(t, one.To, two.To) {
		return false
	}
	if one.TokenID != two.TokenID {
		logger.Info().Msgf("TokenID: %d != %d", one.TokenID, two.TokenID)
		return false
	}
	if one.Amount.Cmp(two.Amount) != 0 {
		logger.Info().Msgf("Amount: %s != %s", one.Amount.String(), two.Amount.String())
		return false
	}
	if one.OperationID != two.OperationID {
		logger.Info().Msgf("OperationID: %d != %d", one.OperationID, two.OperationID)
		return false
	}
	return true
}

func compareOperations(t *testing.T, one, two *operation.Operation) bool {
	if one.Internal != two.Internal {
		logger.Info().Msgf("Internal: %v != %v", one.Internal, two.Internal)
		return false
	}
	if !compareInt64Ptr(one.Nonce, two.Nonce) {
		logger.Info().Msgf("Operation.Nonce: %d != %d", *one.Nonce, *two.Nonce)
		return false
	}
	if one.Timestamp != two.Timestamp {
		logger.Info().Msgf("Timestamp: %s != %s", one.Timestamp, two.Timestamp)
		return false
	}
	if one.Level != two.Level {
		logger.Info().Msgf("Level: %d != %d", one.Level, two.Level)
		return false
	}
	if one.ContentIndex != two.ContentIndex {
		logger.Info().Msgf("ContentIndex: %d != %d", one.ContentIndex, two.ContentIndex)
		return false
	}
	if one.Counter != two.Counter {
		logger.Info().Msgf("Counter: %d != %d", one.Counter, two.Counter)
		return false
	}
	if one.GasLimit != two.GasLimit {
		logger.Info().Msgf("GasLimit: %d != %d", one.GasLimit, two.GasLimit)
		return false
	}
	if one.StorageLimit != two.StorageLimit {
		logger.Info().Msgf("StorageLimit: %d != %d", one.StorageLimit, two.StorageLimit)
		return false
	}
	if one.Fee != two.Fee {
		logger.Info().Msgf("Fee: %d != %d", one.Fee, two.Fee)
		return false
	}
	if one.Amount != two.Amount {
		logger.Info().Msgf("Amount: %d != %d", one.Amount, two.Amount)
		return false
	}
	if one.Burned != two.Burned {
		logger.Info().Msgf("Burned: %d != %d", one.Burned, two.Burned)
		return false
	}
	if one.AllocatedDestinationContractBurned != two.AllocatedDestinationContractBurned {
		logger.Info().Msgf("AllocatedDestinationContractBurned: %d != %d", one.AllocatedDestinationContractBurned, two.AllocatedDestinationContractBurned)
		return false
	}
	if one.ProtocolID != two.ProtocolID {
		logger.Info().Msgf("Protocol: %d != %d", one.ProtocolID, two.ProtocolID)
		return false
	}
	if one.Hash != two.Hash {
		logger.Info().Msgf("Hash: %s != %s", one.Hash, two.Hash)
		return false
	}
	if one.Status != two.Status {
		logger.Info().Msgf("Status: %s != %s", one.Status, two.Status)
		return false
	}
	if one.Kind != two.Kind {
		logger.Info().Msgf("Kind: %s != %s", one.Kind, two.Kind)
		return false
	}
	if !assert.Equal(t, one.Initiator, two.Initiator) {
		return false
	}
	if !assert.Equal(t, one.Source, two.Source) {
		return false
	}
	if !assert.Equal(t, one.Destination, two.Destination) {
		return false
	}
	if !assert.Equal(t, one.Delegate, two.Delegate) {
		return false
	}
	if one.Entrypoint != two.Entrypoint {
		logger.Info().Msgf("Entrypoint: %s != %s", one.Entrypoint, two.Entrypoint)
		return false
	}
	if len(one.Parameters) > 0 && len(two.Parameters) > 0 {
		if !assert.JSONEq(t, string(one.Parameters), string(two.Parameters)) {
			logger.Info().Msgf("Parameters: %s != %s", one.Parameters, two.Parameters)
			return false
		}
	}
	if len(one.DeffatedStorage) > 0 && len(two.DeffatedStorage) > 0 {
		if !assert.JSONEq(t, string(one.DeffatedStorage), string(two.DeffatedStorage)) {
			logger.Info().Msgf("DeffatedStorage: %s != %s", one.DeffatedStorage, two.DeffatedStorage)
			return false
		}
	}
	if one.Tags != two.Tags {
		logger.Info().Msgf("Tags: %d != %d", one.Tags, two.Tags)
		return false
	}

	if len(one.Transfers) != len(two.Transfers) {
		logger.Info().Msgf("Transfers length: %d != %d", len(one.Transfers), len(two.Transfers))
		return false
	}

	if one.Transfers != nil && two.Transfers != nil {
		for i := range one.Transfers {
			if !compareTransfers(t, one.Transfers[i], two.Transfers[i]) {
				return false
			}
		}
	}

	if len(one.BigMapDiffs) != len(two.BigMapDiffs) {
		logger.Info().Msgf("BigMapDiffs length: %d != %d", len(one.BigMapDiffs), len(two.BigMapDiffs))
		return false
	}

	if one.BigMapDiffs != nil && two.BigMapDiffs != nil {
		for i := range one.BigMapDiffs {
			if !compareBigMapDiff(t, one.BigMapDiffs[i], two.BigMapDiffs[i]) {
				return false
			}
		}
	}

	if !assert.Len(t, one.BigMapActions, len(two.BigMapActions)) {
		return false
	}
	if one.BigMapActions != nil && two.BigMapActions != nil {
		for i := range one.BigMapActions {
			if !compareBigMapAction(one.BigMapActions[i], two.BigMapActions[i]) {
				return false
			}
		}
	}

	return true
}

func compareBigMapDiff(t *testing.T, one, two *bigmapdiff.BigMapDiff) bool {
	if one.Contract != two.Contract {
		logger.Info().Msgf("BigMapDiff.Address: %s != %s", one.Contract, two.Contract)
		return false
	}
	if one.KeyHash != two.KeyHash {
		logger.Info().Msgf("KeyHash: %s != %s", one.KeyHash, two.KeyHash)
		return false
	}
	if len(one.Value) > 0 || len(two.Value) > 0 {
		if !assert.JSONEq(t, string(one.ValueBytes()), string(two.ValueBytes())) {
			return false
		}
	}
	if one.Level != two.Level {
		logger.Info().Msgf("Level: %d != %d", one.Level, two.Level)
		return false
	}
	if one.Timestamp != two.Timestamp {
		logger.Info().Msgf("Timestamp: %s != %s", one.Timestamp, two.Timestamp)
		return false
	}
	if one.ProtocolID != two.ProtocolID {
		logger.Info().Msgf("Protocol: %d != %d", one.ProtocolID, two.ProtocolID)
		return false
	}
	if !assert.JSONEq(t, string(one.KeyBytes()), string(two.KeyBytes())) {
		return false
	}
	if len(one.KeyStrings) != len(two.KeyStrings) {
		logger.Info().Msgf("KeyStrings: %v != %v", one.KeyStrings, two.KeyStrings)
		return false
	}
	for i := range one.KeyStrings {
		if one.KeyStrings[i] != two.KeyStrings[i] {
			logger.Info().Msgf("KeyStrings[i]: %v != %v", one.KeyStrings[i], two.KeyStrings[i])
			return false
		}
	}
	return true
}

func compareBigMapAction(one, two *bigmapaction.BigMapAction) bool {
	if one.Action != two.Action {
		logger.Info().Msgf("Action: %s != %s", one.Action, two.Action)
		return false
	}
	if !compareInt64Ptr(one.SourcePtr, two.SourcePtr) {
		logger.Info().Msgf("SourcePtr: %d != %d", *one.SourcePtr, *two.SourcePtr)
		return false
	}
	if !compareInt64Ptr(one.DestinationPtr, two.DestinationPtr) {
		logger.Info().Msgf("DestinationPtr: %d != %d", *one.DestinationPtr, *two.DestinationPtr)
		return false
	}
	if one.Level != two.Level {
		logger.Info().Msgf("Level: %d != %d", one.Level, two.Level)
		return false
	}
	if one.Address != two.Address {
		logger.Info().Msgf("BigMapAction.Address: %s != %s", one.Address, two.Address)
		return false
	}
	if one.Timestamp != two.Timestamp {
		logger.Info().Msgf("Timestamp: %s != %s", one.Timestamp, two.Timestamp)
		return false
	}
	return true
}

func compareContract(t *testing.T, one, two *contract.Contract) bool {
	if !assert.Equal(t, one.Account, two.Account) {
		return false
	}
	if !assert.Equal(t, one.Manager, two.Manager) {
		return false
	}
	if !assert.Equal(t, one.Level, two.Level) {
		return false
	}
	if !assert.Equal(t, one.Timestamp, two.Timestamp) {
		return false
	}
	if !assert.Equal(t, one.Tags, two.Tags) {
		return false
	}
	if !compareScript(t, one.Alpha, two.Alpha) {
		logger.Info().Msgf("Contract.Alpha: %v != %v", one.Alpha, two.Alpha)
		return false
	}
	if !compareScript(t, one.Babylon, two.Babylon) {
		logger.Info().Msgf("Contract.Babylon: %v != %v", one.Babylon, two.Babylon)
		return false
	}
	return true
}

func compareScript(t *testing.T, one, two contract.Script) bool {
	if !assert.Equal(t, one.Hash, two.Hash) {
		return false
	}
	if !assert.ElementsMatch(t, one.Entrypoints, two.Entrypoints) {
		return false
	}
	if !assert.ElementsMatch(t, one.Annotations, two.Annotations) {
		return false
	}
	if !assert.ElementsMatch(t, one.FailStrings, two.FailStrings) {
		return false
	}
	if !assert.ElementsMatch(t, one.Hardcoded, two.Hardcoded) {
		return false
	}
	if !assert.ElementsMatch(t, one.Code, two.Code) {
		return false
	}
	return true
}

func compareInt64Ptr(one, two *int64) bool {
	return (one != nil && two != nil && *one == *two) || (one == nil && two == nil)
}
