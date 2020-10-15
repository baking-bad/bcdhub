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
			if !compareTransfers(t, one, two) {
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
		default:
			log.Printf("Unknown type: %T", one)
			return false
		}
	}

	return true
}

func compareTransfers(t *testing.T, one, two *models.Transfer) bool {
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
	if one.Nonce != nil && two.Nonce != nil && *one.Nonce != *two.Nonce {
		log.Printf("SourcePtr: %d != %d", one.Nonce, two.Nonce)
		return false
	}
	return true
}

func compareOperations(t *testing.T, one, two *models.Operation) bool {
	if one.Internal != two.Internal {
		log.Printf("Internal: %v != %v", one.Internal, two.Internal)
		return false
	}
	if one.Nonce != nil && two.Nonce != nil && *one.Nonce != *two.Nonce {
		log.Printf("SourcePtr: %d != %d", one.Nonce, two.Nonce)
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
	if compareJSON(t, one.Parameters, two.Parameters) {
		log.Printf("Parameters: %s != %s", one.Parameters, two.Parameters)
		return false
	}
	if compareJSON(t, one.DeffatedStorage, two.DeffatedStorage) {
		log.Printf("DeffatedStorage: %s != %s", one.DeffatedStorage, two.DeffatedStorage)
		return false
	}
	if !reflect.DeepEqual(one.ParameterStrings, two.ParameterStrings) {
		log.Printf("ParameterStrings: %s != %s", one.ParameterStrings, two.ParameterStrings)
		return false
	}
	if !reflect.DeepEqual(one.StorageStrings, two.StorageStrings) {
		log.Printf("StorageStrings: %s != %s", one.StorageStrings, two.StorageStrings)
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
		log.Printf("Address: %s != %s", one.Address, two.Address)
		return false
	}
	if one.KeyHash != two.KeyHash {
		log.Printf("KeyHash: %s != %s", one.KeyHash, two.KeyHash)
		return false
	}
	if compareJSON(t, one.Value, two.Value) {
		log.Printf("Value: %s != %s", one.Value, two.Value)
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
	if one.SourcePtr != nil && two.SourcePtr != nil && *one.SourcePtr != *two.SourcePtr {
		log.Printf("SourcePtr: %d != %d", one.SourcePtr, two.SourcePtr)
		return false
	}
	if one.DestinationPtr != nil && two.DestinationPtr != nil && *one.DestinationPtr != *two.DestinationPtr {
		log.Printf("DestinationPtr: %d != %d", one.DestinationPtr, two.DestinationPtr)
		return false
	}
	if one.Level != two.Level {
		log.Printf("Level: %d != %d", one.Level, two.Level)
		return false
	}
	if one.Address != two.Address {
		log.Printf("Address: %s != %s", one.Address, two.Address)
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

func compareJSON(t *testing.T, one, two string) bool {
	if one == "" {
		return one == two
	}
	return assert.JSONEq(t, one, two)
}
