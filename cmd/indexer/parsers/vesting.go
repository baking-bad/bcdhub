package parsers

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/tidwall/gjson"
)

// VestingParser -
type VestingParser struct {
	rpc            noderpc.Pool
	es             *elastic.Elastic
	filesDirectory string
}

// NewVestingParser -
func NewVestingParser(rpc noderpc.Pool, es *elastic.Elastic, filesDirectory string) *VestingParser {
	return &VestingParser{
		rpc:            rpc,
		es:             es,
		filesDirectory: filesDirectory,
	}
}

// Parse -
func (p *VestingParser) Parse(data gjson.Result, head noderpc.Header, network, address string) (models.Migration, *models.Contract, error) {
	migration := models.Migration{
		ID:          helpers.GenerateID(),
		IndexedTime: time.Now().UnixNano() / 1000,

		Level:     head.Level,
		Network:   network,
		Protocol:  head.NextProtocol,
		Address:   address,
		Timestamp: head.Timestamp,
		Kind:      consts.MigrationBootstrap,
	}
	protoSymLink, err := meta.GetProtoSymLink(migration.Protocol)
	if err != nil {
		return migration, nil, err
	}

	op := models.Operation{
		ID:          helpers.GenerateID(),
		Network:     network,
		Protocol:    head.NextProtocol,
		Status:      "applied",
		Kind:        consts.Migration,
		Amount:      data.Get("balance").Int(),
		Counter:     data.Get("counter").Int(),
		Source:      data.Get("manager").String(),
		Destination: address,
		Balance:     data.Get("balance").Int(),
		Delegate:    data.Get("delegate.value").String(),
		Timestamp:   head.Timestamp,
		IndexedTime: time.Now().UnixNano() / 1000,
		Script:      data.Get("script"),
	}
	if !contractparser.IsDelegatorContract(op.Script) {
		contract, err := createNewContract(p.es, op, p.filesDirectory, protoSymLink)
		return migration, contract, err
	}
	return migration, nil, nil
}
