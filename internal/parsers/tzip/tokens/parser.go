package tokens

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/storage"
	"github.com/baking-bad/bcdhub/internal/contractparser/unpack"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/schema"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/tidwall/gjson"
)

// TokenMetadataParser -
type TokenMetadataParser struct {
	bmdRepo      bigmapdiff.Repository
	blocksRepo   block.Repository
	protocolRepo protocol.Repository
	schemaRepo   schema.Repository
	storage      models.GeneralRepository

	rpc       noderpc.INode
	sharePath string
	network   string

	sources map[string]string
}

// NewTokenMetadataParser -
func NewTokenMetadataParser(bmdRepo bigmapdiff.Repository, blocksRepo block.Repository, protocolRepo protocol.Repository, schemaRepo schema.Repository, storage models.GeneralRepository, rpc noderpc.INode, sharePath, network string) TokenMetadataParser {
	return TokenMetadataParser{
		bmdRepo: bmdRepo, blocksRepo: blocksRepo, storage: storage, protocolRepo: protocolRepo, schemaRepo: schemaRepo,
		rpc: rpc, sharePath: sharePath, network: network,
		sources: map[string]string{
			"carthagenet": "tz1grSQDByRpnVs7sPtaprNZRp531ZKz6Jmm",
			"mainnet":     "tz2FCNBrERXtaTtNX6iimR1UJ5JSDxvdHM93",
			"delphinet":   "tz1ME9SBiGDCzLwgoShUMs2d9zRr23aJHf4w",
		},
	}
}

// Parse -
func (t TokenMetadataParser) Parse(address string, level int64) ([]Metadata, error) {
	state, err := t.getState(level)
	if err != nil {
		return nil, err
	}
	registryAddress, err := t.getTokenMetadataRegistry(address, state)
	if err != nil {
		return nil, err
	}
	return t.parse(registryAddress, state)
}

// ParseWithRegistry -
func (t TokenMetadataParser) ParseWithRegistry(registry string, level int64) ([]Metadata, error) {
	state, err := t.getState(level)
	if err != nil {
		return nil, err
	}
	return t.parse(registry, state)
}

func (t TokenMetadataParser) parse(registry string, state block.Block) ([]Metadata, error) {
	ptr, err := t.getBigMapPtr(registry, state)
	if err != nil {
		return nil, err
	}

	bmd, err := t.bmdRepo.Get(bigmapdiff.GetContext{
		Ptr:     &ptr,
		Network: t.network,
		Size:    1000,
	})
	if err != nil {
		return nil, err
	}

	metadata := make([]Metadata, len(bmd))
	for i := range bmd {
		value := gjson.Parse(bmd[i].Value)
		if err := t.parseMetadata(value, &metadata[i]); err != nil {
			continue
		}
		metadata[i].RegistryAddress = registry
		metadata[i].Timestamp = bmd[i].Timestamp
		metadata[i].Level = bmd[i].Level
	}

	return metadata, nil
}

func (t TokenMetadataParser) getState(level int64) (block.Block, error) {
	if level > 0 {
		return t.blocksRepo.GetBlock(t.network, level)
	}
	return t.blocksRepo.GetLastBlock(t.network)
}

func (t TokenMetadataParser) getTokenMetadataRegistry(address string, state block.Block) (string, error) {
	metadata, err := t.hasTokenMetadataRegistry(address, state.Protocol)
	if err != nil {
		return "", err
	} else if metadata == nil {
		return "", ErrNoTokenMetadataRegistryMethod
	}

	source, ok := t.sources[t.network]
	if !ok {
		return "", ErrUnknownNetwork
	}

	result, err := t.storage.SearchByText("view_address", 0, nil, map[string]interface{}{
		"networks": []string{t.network},
		"indices":  []string{models.DocContracts},
	}, false)
	if err != nil {
		return "", err
	}
	if result.Count == 0 {
		return "", ErrNoViewAddressContract
	}

	counter, err := t.rpc.GetCounter(source)
	if err != nil {
		return "", err
	}

	protocol, err := t.protocolRepo.GetProtocol(t.network, "", state.Level)
	if err != nil {
		return "", err
	}

	parameters := gjson.Parse(fmt.Sprintf(`{"entrypoint": "%s", "value": {"string": "%s"}}`, TokenMetadataRegistry, result.Items[0].Value))
	response, err := t.rpc.RunOperation(
		state.ChainID,
		state.Hash,
		source,
		address,
		0,
		protocol.Constants.HardGasLimitPerOperation,
		protocol.Constants.HardStorageLimitPerOperation,
		counter+1,
		0,
		parameters,
	)
	if err != nil {
		return "", err
	}

	registryAddress, err := t.parseRegistryAddress(response)
	if err != nil {
		return "", err
	}
	if !helpers.IsContract(address) {
		return "", ErrInvalidRegistryAddress
	}
	if registryAddress == selfAddress {
		registryAddress = address
	}
	return registryAddress, nil
}

func (t TokenMetadataParser) parseRegistryAddress(response gjson.Result) (string, error) {
	value := response.Get("contents.0.metadata.internal_operation_results.0.parameters.value")
	if value.Exists() {
		if value.Get("bytes").Exists() {
			return unpack.Address(value.Get("bytes").String())
		} else if value.Get("string").Exists() {
			return value.Get("string").String(), nil
		}
	}
	return "", ErrInvalidContractParameter
}

func (t TokenMetadataParser) hasTokenMetadataRegistry(address, protocol string) (meta.Metadata, error) {
	metadata, err := meta.GetMetadata(t.schemaRepo, address, consts.PARAMETER, protocol)
	if err != nil {
		return nil, err
	}

	binPath := metadata.Find(TokenMetadataRegistry)
	if binPath != "" {
		return metadata, nil
	}

	return nil, nil
}

func (t TokenMetadataParser) getBigMapPtr(address string, state block.Block) (int64, error) {
	registryStorageMetadata, err := meta.GetMetadata(t.schemaRepo, address, consts.STORAGE, state.Protocol)
	if err != nil {
		return 0, err
	}

	binPath := registryStorageMetadata.Find(TokenMetadataRegistryStorageKey)
	if binPath == "" {
		return 0, ErrNoMetadataKeyInStorage
	}

	registryStorage, err := t.rpc.GetScriptStorageJSON(address, state.Level)
	if err != nil {
		return 0, err
	}

	ptrs, err := storage.FindBigMapPointers(registryStorageMetadata, registryStorage)
	if err != nil {
		return 0, err
	}
	for ptr, path := range ptrs {
		if path == binPath {
			return ptr, nil
		}
	}

	return 0, ErrNoMetadataKeyInStorage
}

const (
	keyTokenID  = "args.0.int"
	keySymbol   = "args.1.args.0.string"
	keyName     = "args.1.args.1.args.0.string"
	keyDecimals = "args.1.args.1.args.1.args.0.int"
	keyExtras   = "args.1.args.1.args.1.args.1"
)

func (t TokenMetadataParser) parseMetadata(value gjson.Result, m *Metadata) error {
	extras := make(map[string]interface{})
	for _, item := range value.Get(keyExtras).Array() {
		k := item.Get("args.0.string").String()
		if item.Get("args.1.string").Exists() || item.Get("args.1.bytes").Exists() {
			extras[k] = item.Get("args.1.string").String()
		} else if item.Get("args.1.int").Exists() {
			extras[k] = item.Get("args.1.int").Int()
		}
	}

	if !value.Get(keyTokenID).Exists() {
		return ErrInvalidStorageStructure
	}
	if !value.Get(keySymbol).Exists() {
		return ErrInvalidStorageStructure
	}
	if !value.Get(keyName).Exists() {
		return ErrInvalidStorageStructure
	}
	if !value.Get(keyDecimals).Exists() {
		return ErrInvalidStorageStructure
	}

	m.TokenID = value.Get(keyTokenID).Int()
	m.Symbol = value.Get(keySymbol).String()
	m.Name = value.Get(keyName).String()

	val := value.Get(keyDecimals).Int()
	m.Decimals = &val

	m.Extras = extras
	return nil
}
