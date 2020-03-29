package parsers

import (
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/google/uuid"
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
func (p *VestingParser) Parse(data gjson.Result, network, address, protocol string) (models.Migration, *models.Contract, error) {
	ts, err := time.Parse(time.RFC3339, "2018-06-30T00:00:00+00:00")
	if err != nil {
		return models.Migration{}, nil, err
	}
	migration := models.Migration{
		ID:          strings.ReplaceAll(uuid.New().String(), "-", ""),
		IndexedTime: time.Now().UnixNano() / 1000,

		Network:   network,
		Protocol:  protocol,
		Address:   address,
		Timestamp: ts,
		Vesting:   true,
	}
	protoSymLink, err := meta.GetProtoSymLink(migration.Protocol)
	if err != nil {
		return migration, nil, err
	}

	op := models.Operation{
		ID:          strings.ReplaceAll(uuid.New().String(), "-", ""),
		Network:     network,
		Protocol:    protocol,
		Status:      "applied",
		Kind:        consts.Migration,
		Amount:      data.Get("balance").Int(),
		Counter:     data.Get("counter").Int(),
		Source:      data.Get("manager").String(),
		Destination: address,
		Balance:     data.Get("balance").Int(),
		Delegate:    data.Get("delegate.value").String(),
		Timestamp:   ts,
		IndexedTime: time.Now().UnixNano() / 1000,
		Script:      data.Get("script"),
	}
	if !contractparser.IsDelegateContract(op.Script) {
		contract, err := createNewContract(p.es, op, p.filesDirectory, protoSymLink)
		return migration, contract, err
	}
	return migration, nil, nil
}
