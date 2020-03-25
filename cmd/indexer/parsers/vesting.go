package parsers

import (
	"fmt"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
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
	protocols      map[string]string
}

// NewVestingParser -
func NewVestingParser(rpc noderpc.Pool, es *elastic.Elastic, filesDirectory string, protocols map[string]string) *VestingParser {
	return &VestingParser{
		rpc:            rpc,
		es:             es,
		filesDirectory: filesDirectory,
		protocols:      protocols,
	}
}

// Parse -
func (p *VestingParser) Parse(data gjson.Result, network, address, protocol string) (models.Operation, *models.Contract, error) {
	ts, err := time.Parse(time.RFC3339, "2018-06-30T00:00:00+00:00")
	if err != nil {
		return models.Operation{}, nil, err
	}
	op := models.Operation{
		ID:          strings.ReplaceAll(uuid.New().String(), "-", ""),
		Network:     network,
		Protocol:    protocol,
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

	protoSymLink, ok := p.protocols[protocol]
	if !ok {
		return op, nil, fmt.Errorf("[%s] Unknown protocol: %s", network, protocol)
	}
	if !contractparser.IsDelegateContract(op.Script) {
		contract, err := createNewContract(p.es, op, p.filesDirectory, protoSymLink)
		return op, contract, err
	}
	return op, nil, nil
}
