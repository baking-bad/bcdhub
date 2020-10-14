package contract

import (
	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/kinds"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
)

// Parser -
type Parser struct {
	interfaces     map[string]kinds.ContractKind
	filesDirectory string

	metadata map[string]*meta.ContractMetadata

	scriptSaver ScriptSaver
}

// NewParser -
func NewParser(interfaces map[string]kinds.ContractKind, opts ...ParserOption) Parser {
	parser := Parser{
		interfaces: interfaces,
		metadata:   make(map[string]*meta.ContractMetadata),
	}

	for i := range opts {
		opts[i](&parser)
	}

	return parser
}

// ParserOption -
type ParserOption func(p *Parser)

// WithShareDirContractParser -
func WithShareDirContractParser(dir string) ParserOption {
	return func(p *Parser) {
		if dir == "" {
			return
		}
		p.filesDirectory = dir
		p.scriptSaver = NewFileScriptSaver(dir)
	}
}

// Parse -
func (p *Parser) Parse(operation models.Operation) ([]elastic.Model, error) {
	if !helpers.StringInArray(operation.Kind, []string{
		consts.Origination, consts.OriginationNew, consts.Migration,
	}) {
		return nil, errors.Errorf("Invalid operation kind in computeContractMetrics: %s", operation.Kind)
	}
	contract := models.Contract{
		ID:         helpers.GenerateID(),
		Network:    operation.Network,
		Level:      operation.Level,
		Timestamp:  operation.Timestamp,
		Manager:    operation.Source,
		Address:    operation.Destination,
		Balance:    operation.Amount,
		Delegate:   operation.Delegate,
		LastAction: models.BCDTime{Time: operation.Timestamp},
		TxCount:    1,
	}

	protoSymLink, err := meta.GetProtoSymLink(operation.Protocol)
	if err != nil {
		return nil, err
	}

	if err := p.computeMetrics(operation, protoSymLink, &contract); err != nil {
		return nil, err
	}

	metadata, err := NewMetadataParser(protoSymLink).Parse(operation.Script, contract.Address)
	if err != nil {
		return nil, err
	}

	contractMetadata, err := meta.GetContractMetadataFromModel(metadata)
	if err != nil {
		return nil, err
	}
	p.metadata[metadata.ID] = contractMetadata

	if contractMetadata.IsUpgradable(protoSymLink) {
		contract.Tags = append(contract.Tags, consts.UpgradableTag)
	}

	if err := setEntrypoints(contractMetadata, protoSymLink, &contract); err != nil {
		return nil, err
	}

	return []elastic.Model{&metadata, &contract}, nil
}

// GetContractMetadata -
func (p Parser) GetContractMetadata(address string) (*meta.ContractMetadata, error) {
	metadata, ok := p.metadata[address]
	if !ok {
		return nil, errors.Errorf("Unknown parsed metadata: %s", address)
	}
	return metadata, nil
}

func (p Parser) computeMetrics(operation models.Operation, protoSymLink string, contract *models.Contract) error {
	script, err := contractparser.New(operation.Script)
	if err != nil {
		return errors.Errorf("contractparser.New: %v", err)
	}
	script.Parse(p.interfaces)

	lang, err := script.Language()
	if err != nil {
		return errors.Errorf("script.Language: %v", err)
	}

	contract.Language = lang
	contract.Hash = script.Code.Hash
	contract.FailStrings = script.Code.FailStrings.Values()
	contract.Annotations = script.Annotations.Values()
	contract.Tags = script.Tags.Values()
	contract.Hardcoded = script.HardcodedAddresses.Values()

	if err := metrics.SetFingerprint(operation.Script, contract); err != nil {
		return err
	}
	if p.scriptSaver != nil {
		return p.scriptSaver.Save(operation.Script, scriptSaveContext{
			Network: contract.Network,
			Address: contract.Address,
			SymLink: protoSymLink,
		})
	}
	return nil
}

func setEntrypoints(metadata *meta.ContractMetadata, symLink string, contract *models.Contract) error {
	entrypoints, err := metadata.Parameter[symLink].GetEntrypoints()
	if err != nil {
		return err
	}
	contract.Entrypoints = make([]string, len(entrypoints))
	for i := range entrypoints {
		contract.Entrypoints[i] = entrypoints[i].Name
	}
	return nil
}
