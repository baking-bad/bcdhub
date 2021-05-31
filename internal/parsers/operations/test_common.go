package operations

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func newDecimal(val string) decimal.Decimal {
	i, _ := decimal.NewFromString(val)
	return i
}

func readJSONFile(name string, response interface{}) error {
	bytes, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, response)
}

func readTestContractModel(network types.Network, address string) (contract.Contract, error) {
	var c contract.Contract
	bytes, err := ioutil.ReadFile(fmt.Sprintf("./data/models/contract/%s.json", address))
	if err != nil {
		return c, err
	}
	err = json.Unmarshal(bytes, &c)
	return c, err
}

func readStorage(address string, level int64) ([]byte, error) {
	storageFile := fmt.Sprintf("./data/rpc/script/storage/%s_%d.json", address, level)
	return ioutil.ReadFile(storageFile)
}

func compareParserResponse(t *testing.T, got, want *parsers.Result) bool {
	assert.Len(t, got.BigMapActions, len(want.BigMapActions))
	assert.Len(t, got.BigMapDiffs, len(want.BigMapDiffs))
	assert.Len(t, got.BigMapState, len(want.BigMapState))
	assert.Len(t, got.Contracts, len(want.Contracts))
	assert.Len(t, got.Migrations, len(want.Migrations))
	assert.Len(t, got.Operations, len(want.Operations))
	assert.Len(t, got.TokenBalances, len(want.TokenBalances))
	assert.Len(t, got.Transfers, len(want.Transfers))

	for i := range got.BigMapActions {
		if !compareBigMapAction(want.BigMapActions[i], got.BigMapActions[i]) {
			return false
		}
	}
	for i := range got.BigMapDiffs {
		if !compareBigMapDiff(t, want.BigMapDiffs[i], got.BigMapDiffs[i]) {
			return false
		}
	}
	for i := range got.BigMapState {
		if !assert.Equal(t, want.BigMapState[i], got.BigMapState[i]) {
			return false
		}
	}
	for i := range got.Contracts {
		if !compareContract(want.Contracts[i], got.Contracts[i]) {
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
	for i := range got.Transfers {
		if !compareTransfers(want.Transfers[i], got.Transfers[i]) {
			return false
		}
	}

	return true
}

func compareTransfers(one, two *transfer.Transfer) bool {
	if one.Network != two.Network {
		logger.Info("Network: %s != %s", one.Network, two.Network)
		return false
	}
	if one.Contract != two.Contract {
		logger.Info("Contract: %s != %s", one.Contract, two.Contract)
		return false
	}
	if one.Initiator != two.Initiator {
		logger.Info("Initiator: %s != %s", one.Initiator, two.Initiator)
		return false
	}
	if one.Hash != two.Hash {
		logger.Info("Hash: %s != %s", one.Hash, two.Hash)
		return false
	}
	if one.Status != two.Status {
		logger.Info("Status: %s != %s", one.Status, two.Status)
		return false
	}
	if one.Timestamp != two.Timestamp {
		logger.Info("Timestamp: %s != %s", one.Timestamp, two.Timestamp)
		return false
	}
	if one.Level != two.Level {
		logger.Info("Level: %d != %d", one.Level, two.Level)
		return false
	}
	if one.From != two.From {
		logger.Info("From: %s != %s", one.From, two.From)
		return false
	}
	if one.To != two.To {
		logger.Info("To: %s != %s", one.To, two.To)
		return false
	}
	if one.TokenID != two.TokenID {
		logger.Info("TokenID: %d != %d", one.TokenID, two.TokenID)
		return false
	}
	if one.Amount.Cmp(two.Amount) != 0 {
		logger.Info("Amount: %s != %s", one.Amount.String(), two.Amount.String())
		return false
	}
	if one.Counter != two.Counter {
		logger.Info("Counter: %d != %d", one.Counter, two.Counter)
		return false
	}
	if !compareInt64Ptr(one.Nonce, two.Nonce) {
		logger.Info("Transfer.Nonce: %d != %d", *one.Nonce, *two.Nonce)
		return false
	}
	return true
}

func compareOperations(t *testing.T, one, two *operation.Operation) bool {
	if one.Internal != two.Internal {
		logger.Info("Internal: %v != %v", one.Internal, two.Internal)
		return false
	}
	if !compareInt64Ptr(one.Nonce, two.Nonce) {
		logger.Info("Operation.Nonce: %d != %d", *one.Nonce, *two.Nonce)
		return false
	}
	if one.Timestamp != two.Timestamp {
		logger.Info("Timestamp: %s != %s", one.Timestamp, two.Timestamp)
		return false
	}
	if one.Level != two.Level {
		logger.Info("Level: %d != %d", one.Level, two.Level)
		return false
	}
	if one.ContentIndex != two.ContentIndex {
		logger.Info("ContentIndex: %d != %d", one.ContentIndex, two.ContentIndex)
		return false
	}
	if one.Counter != two.Counter {
		logger.Info("Counter: %d != %d", one.Counter, two.Counter)
		return false
	}
	if one.GasLimit != two.GasLimit {
		logger.Info("GasLimit: %d != %d", one.GasLimit, two.GasLimit)
		return false
	}
	if one.StorageLimit != two.StorageLimit {
		logger.Info("StorageLimit: %d != %d", one.StorageLimit, two.StorageLimit)
		return false
	}
	if one.Fee != two.Fee {
		logger.Info("Fee: %d != %d", one.Fee, two.Fee)
		return false
	}
	if one.Amount != two.Amount {
		logger.Info("Amount: %d != %d", one.Amount, two.Amount)
		return false
	}
	if one.Burned != two.Burned {
		logger.Info("Burned: %d != %d", one.Burned, two.Burned)
		return false
	}
	if one.AllocatedDestinationContractBurned != two.AllocatedDestinationContractBurned {
		logger.Info("AllocatedDestinationContractBurned: %d != %d", one.AllocatedDestinationContractBurned, two.AllocatedDestinationContractBurned)
		return false
	}
	if one.Network != two.Network {
		logger.Info("Network: %s != %s", one.Network, two.Network)
		return false
	}
	if one.Protocol != two.Protocol {
		logger.Info("Protocol: %s != %s", one.Protocol, two.Protocol)
		return false
	}
	if one.Hash != two.Hash {
		logger.Info("Hash: %s != %s", one.Hash, two.Hash)
		return false
	}
	if one.Status != two.Status {
		logger.Info("Status: %s != %s", one.Status, two.Status)
		return false
	}
	if one.Kind != two.Kind {
		logger.Info("Kind: %s != %s", one.Kind, two.Kind)
		return false
	}
	if one.Initiator != two.Initiator {
		logger.Info("Initiator: %s != %s", one.Initiator, two.Initiator)
		return false
	}
	if one.Source != two.Source {
		logger.Info("Source: %s != %s", one.Source, two.Source)
		return false
	}
	if one.Destination != two.Destination {
		logger.Info("Destination: %s != %s", one.Destination, two.Destination)
		return false
	}
	if one.Delegate != two.Delegate {
		logger.Info("Delegate: %s != %s", one.Delegate, two.Delegate)
		return false
	}
	if one.Entrypoint != two.Entrypoint {
		logger.Info("Entrypoint: %s != %s", one.Entrypoint, two.Entrypoint)
		return false
	}
	if one.SourceAlias != two.SourceAlias {
		logger.Info("SourceAlias: %s != %s", one.SourceAlias, two.SourceAlias)
		return false
	}
	if one.DestinationAlias != two.DestinationAlias {
		logger.Info("DestinationAlias: %s != %s", one.DestinationAlias, two.DestinationAlias)
		return false
	}
	if one.DelegateAlias != two.DelegateAlias {
		logger.Info("DelegateAlias: %s != %s", one.DelegateAlias, two.DelegateAlias)
		return false
	}
	if len(one.Parameters) > 0 && len(two.Parameters) > 0 {
		if !assert.JSONEq(t, string(one.Parameters), string(two.Parameters)) {
			logger.Info("Parameters: %s != %s", one.Parameters, two.Parameters)
			return false
		}
	}
	if len(one.DeffatedStorage) > 0 && len(two.DeffatedStorage) > 0 {
		if !assert.JSONEq(t, string(one.DeffatedStorage), string(two.DeffatedStorage)) {
			logger.Info("DeffatedStorage: %s != %s", one.DeffatedStorage, two.DeffatedStorage)
			return false
		}
	}
	if len(one.Tags) == len(two.Tags) && len(one.Tags) > 0 {
		if !reflect.DeepEqual(one.Tags, two.Tags) {
			logger.Info("Tags: %s != %s", one.Tags, two.Tags)
			return false
		}
	}
	return true
}

func compareBigMapDiff(t *testing.T, one, two *bigmapdiff.BigMapDiff) bool {
	if one.Contract != two.Contract {
		logger.Info("BigMapDiff.Address: %s != %s", one.Contract, two.Contract)
		return false
	}
	if one.KeyHash != two.KeyHash {
		logger.Info("KeyHash: %s != %s", one.KeyHash, two.KeyHash)
		return false
	}
	if len(one.Value) > 0 || len(two.Value) > 0 {
		if !assert.JSONEq(t, string(one.ValueBytes()), string(two.ValueBytes())) {
			return false
		}
	}
	if one.Level != two.Level {
		logger.Info("Level: %d != %d", one.Level, two.Level)
		return false
	}
	if one.Network != two.Network {
		logger.Info("Network: %s != %s", one.Network, two.Network)
		return false
	}
	if one.Timestamp != two.Timestamp {
		logger.Info("Timestamp: %s != %s", one.Timestamp, two.Timestamp)
		return false
	}
	if one.Protocol != two.Protocol {
		logger.Info("Protocol: %s != %s", one.Protocol, two.Protocol)
		return false
	}
	if !assert.JSONEq(t, string(one.KeyBytes()), string(two.KeyBytes())) {
		return false
	}
	return true
}

func compareBigMapAction(one, two *bigmapaction.BigMapAction) bool {
	if one.Action != two.Action {
		logger.Info("Action: %s != %s", one.Action, two.Action)
		return false
	}
	if !compareInt64Ptr(one.SourcePtr, two.SourcePtr) {
		logger.Info("SourcePtr: %d != %d", *one.SourcePtr, *two.SourcePtr)
		return false
	}
	if !compareInt64Ptr(one.DestinationPtr, two.DestinationPtr) {
		logger.Info("DestinationPtr: %d != %d", *one.DestinationPtr, *two.DestinationPtr)
		return false
	}
	if one.Level != two.Level {
		logger.Info("Level: %d != %d", one.Level, two.Level)
		return false
	}
	if one.Address != two.Address {
		logger.Info("BigMapAction.Address: %s != %s", one.Address, two.Address)
		return false
	}
	if one.Network != two.Network {
		logger.Info("Network: %s != %s", one.Network, two.Network)
		return false
	}
	if one.Timestamp != two.Timestamp {
		logger.Info("Timestamp: %s != %s", one.Timestamp, two.Timestamp)
		return false
	}
	return true
}

func compareContract(one, two *contract.Contract) bool {
	if one.Network != two.Network {
		logger.Info("Contract.Network: %s != %s", one.Network, two.Network)
		return false
	}
	if one.Address != two.Address {
		logger.Info("Contract.Address: %s != %s", one.Address, two.Address)
		return false
	}
	if one.Language != two.Language {
		logger.Info("Contract.Language: %s != %s", one.Language, two.Language)
		return false
	}
	if one.Hash != two.Hash {
		logger.Info("Contract.Hash: %s != %s", one.Hash, two.Hash)
		return false
	}
	if one.Manager != two.Manager {
		logger.Info("Contract.Manager: %s != %s", one.Manager, two.Manager)
		return false
	}
	if one.Level != two.Level {
		logger.Info("Contract.Level: %d != %d", one.Level, two.Level)
		return false
	}
	if one.Timestamp != two.Timestamp {
		logger.Info("Contract.Timestamp: %s != %s", one.Timestamp, two.Timestamp)
		return false
	}
	if !compareStringArray(one.Tags, two.Tags) {
		logger.Info("Contract.Tags: %v != %v", one.Tags, two.Tags)
		return false
	}
	if !compareStringArray(one.Entrypoints, two.Entrypoints) {
		logger.Info("Contract.Entrypoints: %v != %v", one.Entrypoints, two.Entrypoints)
		return false
	}
	return true
}

func compareInt64Ptr(one, two *int64) bool {
	return (one != nil && two != nil && *one == *two) || (one == nil && two == nil)
}

func compareStringArray(one, two []string) bool {
	if len(one) != len(two) {
		return false
	}

	for i := range one {
		if !helpers.StringInArray(one[i], two) {
			return false
		}
	}

	return true
}
