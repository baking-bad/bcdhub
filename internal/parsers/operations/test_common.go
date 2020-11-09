package operations

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"testing"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func readJSONFile(name string) (gjson.Result, error) {
	bytes, err := ioutil.ReadFile(name)
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.ParseBytes(bytes), nil
}

func readTestMetadata(address string) (*meta.ContractMetadata, error) {
	bytes, err := ioutil.ReadFile(fmt.Sprintf("./data/metadata/%s.json", address))
	if err != nil {
		return nil, err
	}
	var metadata meta.ContractMetadata
	err = json.Unmarshal(bytes, &metadata)
	return &metadata, err
}

func readTestMetadataModel(address string) (*models.Metadata, error) {
	bytes, err := ioutil.ReadFile(fmt.Sprintf("./data/models/metadata/%s.json", address))
	if err != nil {
		return nil, err
	}
	var metadata models.Metadata
	err = json.Unmarshal(bytes, &metadata)
	return &metadata, err
}

func readTestContractModel(address string) (models.Contract, error) {
	bytes, err := ioutil.ReadFile(fmt.Sprintf("./data/models/contract/%s.json", address))
	if err != nil {
		return models.Contract{}, err
	}
	var contract models.Contract
	err = json.Unmarshal(bytes, &contract)
	return contract, err
}

func readStorage(address string, level int64) (gjson.Result, error) {
	storageFile := fmt.Sprintf("./data/rpc/script/storage/%s_%d.json", address, level)
	return readJSONFile(storageFile)
}

func compareParserResponse(t *testing.T, got, want []elastic.Model) bool {
	if len(got) != len(want) {
		log.Printf("len(got) != len(want): %d != %d", len(got), len(want))
		return false
	}
	for i := range got {
		switch one := got[i].(type) {
		case *models.Transfer:
			two, ok := want[i].(*models.Transfer)
			if !ok {
				log.Printf("Differrrent types: %T != %T", one, two)
				return false
			}
			if !compareTransfers(one, two) {
				return false
			}
		case *models.Operation:
			two, ok := want[i].(*models.Operation)
			if !ok {
				log.Printf("Differrrent types: %T != %T", one, two)
				return false
			}
			if !compareOperations(t, one, two) {
				return false
			}
		case *models.BigMapDiff:
			two, ok := want[i].(*models.BigMapDiff)
			if !ok {
				log.Printf("Differrrent types: %T != %T", one, two)
				return false
			}
			if !compareBigMapDiff(t, one, two) {
				return false
			}
		case *models.BigMapAction:
			two, ok := want[i].(*models.BigMapAction)
			if !ok {
				return false
			}
			if !compareBigMapAction(one, two) {
				return false
			}
		case *models.Contract:
			two, ok := want[i].(*models.Contract)
			if !ok {
				return false
			}
			if !compareContract(one, two) {
				return false
			}
		case *models.Metadata:
			two, ok := want[i].(*models.Metadata)
			if !ok {
				return false
			}
			if !compareMetadata(t, one, two) {
				return false
			}
		case *models.BalanceUpdate:
			two, ok := want[i].(*models.BalanceUpdate)
			if !ok {
				return false
			}
			if !compareBalanceUpdates(one, two) {
				return false
			}
		default:
			log.Printf("Unknown type: %T", one)
			return false
		}
	}

	return true
}

func compareTransfers(one, two *models.Transfer) bool {
	if one.Network != two.Network {
		log.Printf("Network: %s != %s", one.Network, two.Network)
		return false
	}
	if one.Contract != two.Contract {
		log.Printf("Contract: %s != %s", one.Contract, two.Contract)
		return false
	}
	if one.Alias != two.Alias {
		log.Printf("Alias: %s != %s", one.Alias, two.Alias)
		return false
	}
	if one.Initiator != two.Initiator {
		log.Printf("Initiator: %s != %s", one.Initiator, two.Initiator)
		return false
	}
	if one.InitiatorAlias != two.InitiatorAlias {
		log.Printf("InitiatorAlias: %s != %s", one.InitiatorAlias, two.InitiatorAlias)
		return false
	}
	if one.Hash != two.Hash {
		log.Printf("Hash: %s != %s", one.Hash, two.Hash)
		return false
	}
	if one.Status != two.Status {
		log.Printf("Status: %s != %s", one.Status, two.Status)
		return false
	}
	if one.Timestamp != two.Timestamp {
		log.Printf("Timestamp: %s != %s", one.Timestamp, two.Timestamp)
		return false
	}
	if one.Level != two.Level {
		log.Printf("Level: %d != %d", one.Level, two.Level)
		return false
	}
	if one.From != two.From {
		log.Printf("From: %s != %s", one.From, two.From)
		return false
	}
	if one.FromAlias != two.FromAlias {
		log.Printf("FromAlias: %s != %s", one.FromAlias, two.FromAlias)
		return false
	}
	if one.To != two.To {
		log.Printf("To: %s != %s", one.To, two.To)
		return false
	}
	if one.ToAlias != two.ToAlias {
		log.Printf("ToAlias: %s != %s", one.ToAlias, two.ToAlias)
		return false
	}
	if one.TokenID != two.TokenID {
		log.Printf("TokenID: %d != %d", one.TokenID, two.TokenID)
		return false
	}
	if one.Amount != two.Amount {
		log.Printf("Amount: %f != %f", one.Amount, two.Amount)
		return false
	}
	if one.Counter != two.Counter {
		log.Printf("Counter: %d != %d", one.Counter, two.Counter)
		return false
	}
	if !compareInt64Ptr(one.Nonce, two.Nonce) {
		log.Printf("Transfer.Nonce: %d != %d", *one.Nonce, *two.Nonce)
		return false
	}
	return true
}

func compareOperations(t *testing.T, one, two *models.Operation) bool {
	if one.Internal != two.Internal {
		log.Printf("Internal: %v != %v", one.Internal, two.Internal)
		return false
	}
	if !compareInt64Ptr(one.Nonce, two.Nonce) {
		log.Printf("Operation.Nonce: %d != %d", *one.Nonce, *two.Nonce)
		return false
	}
	if one.Timestamp != two.Timestamp {
		log.Printf("Timestamp: %s != %s", one.Timestamp, two.Timestamp)
		return false
	}
	if one.Level != two.Level {
		log.Printf("Level: %d != %d", one.Level, two.Level)
		return false
	}
	if one.ContentIndex != two.ContentIndex {
		log.Printf("ContentIndex: %d != %d", one.ContentIndex, two.ContentIndex)
		return false
	}
	if one.Counter != two.Counter {
		log.Printf("Counter: %d != %d", one.Counter, two.Counter)
		return false
	}
	if one.GasLimit != two.GasLimit {
		log.Printf("GasLimit: %d != %d", one.GasLimit, two.GasLimit)
		return false
	}
	if one.StorageLimit != two.StorageLimit {
		log.Printf("StorageLimit: %d != %d", one.StorageLimit, two.StorageLimit)
		return false
	}
	if one.Fee != two.Fee {
		log.Printf("Fee: %d != %d", one.Fee, two.Fee)
		return false
	}
	if one.Amount != two.Amount {
		log.Printf("Amount: %d != %d", one.Amount, two.Amount)
		return false
	}
	if one.Burned != two.Burned {
		log.Printf("Burned: %d != %d", one.Burned, two.Burned)
		return false
	}
	if one.AllocatedDestinationContractBurned != two.AllocatedDestinationContractBurned {
		log.Printf("AllocatedDestinationContractBurned: %d != %d", one.AllocatedDestinationContractBurned, two.AllocatedDestinationContractBurned)
		return false
	}
	if one.Network != two.Network {
		log.Printf("Network: %s != %s", one.Network, two.Network)
		return false
	}
	if one.Protocol != two.Protocol {
		log.Printf("Protocol: %s != %s", one.Protocol, two.Protocol)
		return false
	}
	if one.Hash != two.Hash {
		log.Printf("Hash: %s != %s", one.Hash, two.Hash)
		return false
	}
	if one.Status != two.Status {
		log.Printf("Status: %s != %s", one.Status, two.Status)
		return false
	}
	if one.Kind != two.Kind {
		log.Printf("Kind: %s != %s", one.Kind, two.Kind)
		return false
	}
	if one.Initiator != two.Initiator {
		log.Printf("Initiator: %s != %s", one.Initiator, two.Initiator)
		return false
	}
	if one.Source != two.Source {
		log.Printf("Source: %s != %s", one.Source, two.Source)
		return false
	}
	if one.Destination != two.Destination {
		log.Printf("Destination: %s != %s", one.Destination, two.Destination)
		return false
	}
	if one.PublicKey != two.PublicKey {
		log.Printf("PublicKey: %s != %s", one.PublicKey, two.PublicKey)
		return false
	}
	if one.ManagerPubKey != two.ManagerPubKey {
		log.Printf("ManagerPubKey: %s != %s", one.ManagerPubKey, two.ManagerPubKey)
		return false
	}
	if one.Delegate != two.Delegate {
		log.Printf("Delegate: %s != %s", one.Delegate, two.Delegate)
		return false
	}
	if one.Entrypoint != two.Entrypoint {
		log.Printf("Entrypoint: %s != %s", one.Entrypoint, two.Entrypoint)
		return false
	}
	if one.SourceAlias != two.SourceAlias {
		log.Printf("SourceAlias: %s != %s", one.SourceAlias, two.SourceAlias)
		return false
	}
	if one.DestinationAlias != two.DestinationAlias {
		log.Printf("DestinationAlias: %s != %s", one.DestinationAlias, two.DestinationAlias)
		return false
	}
	if one.DelegateAlias != two.DelegateAlias {
		log.Printf("DelegateAlias: %s != %s", one.DelegateAlias, two.DelegateAlias)
		return false
	}
	if !compareJSON(t, one.Parameters, two.Parameters) {
		log.Printf("Parameters: %s != %s", one.Parameters, two.Parameters)
		return false
	}
	if !compareJSON(t, one.DeffatedStorage, two.DeffatedStorage) {
		log.Printf("DeffatedStorage: %s != %s", one.DeffatedStorage, two.DeffatedStorage)
		return false
	}
	if !reflect.DeepEqual(one.Tags, two.Tags) {
		log.Printf("Tags: %s != %s", one.Tags, two.Tags)
		return false
	}
	return true
}

func compareBigMapDiff(t *testing.T, one, two *models.BigMapDiff) bool {
	if one.Address != two.Address {
		log.Printf("BigMapDiff.Address: %s != %s", one.Address, two.Address)
		return false
	}
	if one.KeyHash != two.KeyHash {
		log.Printf("KeyHash: %s != %s", one.KeyHash, two.KeyHash)
		return false
	}
	if !compareJSON(t, one.Value, two.Value) {
		log.Printf("BigMapDiff.Value: %s != %s", one.Value, two.Value)
		return false
	}
	if one.BinPath != two.BinPath {
		log.Printf("BinPath: %s != %s", one.BinPath, two.BinPath)
		return false
	}
	if one.Level != two.Level {
		log.Printf("Level: %d != %d", one.Level, two.Level)
		return false
	}
	if one.Network != two.Network {
		log.Printf("Network: %s != %s", one.Network, two.Network)
		return false
	}
	if one.Timestamp != two.Timestamp {
		log.Printf("Timestamp: %s != %s", one.Timestamp, two.Timestamp)
		return false
	}
	if one.Protocol != two.Protocol {
		log.Printf("Protocol: %s != %s", one.Protocol, two.Protocol)
		return false
	}
	if !reflect.DeepEqual(one.Key, two.Key) {
		log.Printf("Key: %s != %s", one.Key, two.Key)
		return false
	}
	return true
}

func compareBigMapAction(one, two *models.BigMapAction) bool {
	if one.Action != two.Action {
		log.Printf("Action: %s != %s", one.Action, two.Action)
		return false
	}
	if !compareInt64Ptr(one.SourcePtr, two.SourcePtr) {
		log.Printf("SourcePtr: %d != %d", *one.SourcePtr, *two.SourcePtr)
		return false
	}
	if !compareInt64Ptr(one.DestinationPtr, two.DestinationPtr) {
		log.Printf("DestinationPtr: %d != %d", *one.DestinationPtr, *two.DestinationPtr)
		return false
	}
	if one.Level != two.Level {
		log.Printf("Level: %d != %d", one.Level, two.Level)
		return false
	}
	if one.Address != two.Address {
		log.Printf("BigMapAction.Address: %s != %s", one.Address, two.Address)
		return false
	}
	if one.Network != two.Network {
		log.Printf("Network: %s != %s", one.Network, two.Network)
		return false
	}
	if one.Timestamp != two.Timestamp {
		log.Printf("Timestamp: %s != %s", one.Timestamp, two.Timestamp)
		return false
	}
	return true
}

func compareContract(one, two *models.Contract) bool {
	if one.Network != two.Network {
		log.Printf("Contract.Network: %s != %s", one.Network, two.Network)
		return false
	}
	if one.Address != two.Address {
		log.Printf("Contract.Address: %s != %s", one.Address, two.Address)
		return false
	}
	if one.Language != two.Language {
		log.Printf("Contract.Language: %s != %s", one.Language, two.Language)
		return false
	}
	if one.Hash != two.Hash {
		log.Printf("Contract.Hash: %s != %s", one.Hash, two.Hash)
		return false
	}
	if one.Manager != two.Manager {
		log.Printf("Contract.Manager: %s != %s", one.Manager, two.Manager)
		return false
	}
	if one.Level != two.Level {
		log.Printf("Contract.Level: %d != %d", one.Level, two.Level)
		return false
	}
	if one.Timestamp != two.Timestamp {
		log.Printf("Contract.Timestamp: %s != %s", one.Timestamp, two.Timestamp)
		return false
	}
	if !compareStringArray(one.Tags, two.Tags) {
		log.Printf("Contract.Tags: %v != %v", one.Tags, two.Tags)
		return false
	}
	if !compareStringArray(one.Hardcoded, two.Hardcoded) {
		log.Printf("Contract.Hardcoded: %v != %v", one.Hardcoded, two.Hardcoded)
		return false
	}
	if !compareStringArray(one.FailStrings, two.FailStrings) {
		log.Printf("Contract.FailStrings: %v != %v", one.FailStrings, two.FailStrings)
		return false
	}
	if !compareStringArray(one.Annotations, two.Annotations) {
		log.Printf("Contract.Annotations: %v != %v", one.Annotations, two.Annotations)
		return false
	}
	if !compareStringArray(one.Entrypoints, two.Entrypoints) {
		log.Printf("Contract.Entrypoints: %v != %v", one.Entrypoints, two.Entrypoints)
		return false
	}
	return true
}

func compareBalanceUpdates(a, b *models.BalanceUpdate) bool {
	if a.Change != b.Change {
		log.Printf("BalanceUpdate.Change: %d != %d", a.Change, b.Change)
		return false
	}
	if a.Contract != b.Contract {
		log.Printf("BalanceUpdate.Contract: %s != %s", a.Contract, b.Contract)
		return false
	}
	if a.Network != b.Network {
		log.Printf("BalanceUpdate.Network: %s != %s", a.Network, b.Network)
		return false
	}
	if a.Level != b.Level {
		log.Printf("BalanceUpdate.Level: %d != %d", a.Level, b.Level)
		return false
	}
	if a.OperationHash != b.OperationHash {
		log.Printf("BalanceUpdate.OperationHash: %s != %s", a.OperationHash, b.OperationHash)
		return false
	}
	if a.ContentIndex != b.ContentIndex {
		log.Printf("BalanceUpdate.ContentIndex: %d != %d", a.ContentIndex, b.ContentIndex)
		return false
	}
	if !compareInt64Ptr(a.Nonce, b.Nonce) {
		log.Printf("BalanceUpdate.Nonce: %d != %d", *a.Nonce, *b.Nonce)
		return false
	}
	return true
}

func compareMetadata(t *testing.T, one, two *models.Metadata) bool {
	if one.ID != two.ID {
		log.Printf("Metadata.ID: %s != %s", one.ID, two.ID)
		return false
	}

	for key, value := range one.Parameter {
		if valueTwo, ok := two.Parameter[key]; ok {
			if !compareJSON(t, value, valueTwo) {
				log.Printf("Metadata.Parameter[%s]: %s != %s", key, value, valueTwo)
				return false
			}
		} else {
			log.Printf("Metadata.Parameter[%s] is absent", key)
			return false
		}
	}

	for key, value := range one.Storage {
		if valueTwo, ok := two.Storage[key]; ok {
			if !compareJSON(t, value, valueTwo) {
				log.Printf("Metadata.Storage[%s]: %s != %s", key, value, valueTwo)
				return false
			}
		} else {
			log.Printf("Metadata.Storage[%s] is absent", key)
			return false
		}
	}
	return true
}

func compareJSON(t *testing.T, one, two string) bool {
	if one == "" {
		return one == two
	}
	return assert.JSONEq(t, one, two)
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
