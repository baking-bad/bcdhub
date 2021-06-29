package handlers

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

const (
	recordsAnnot   = "records"
	expiryMapAnnot = "expiry_map"
)

// TezosDomain -
type TezosDomain struct {
	storage models.GeneralRepository

	contracts map[contract.Address]struct{}
}

// NewTezosDomains -
func NewTezosDomains(storage models.GeneralRepository, contracts map[types.Network]string) *TezosDomain {
	addresses := make(map[contract.Address]struct{})
	for k, v := range contracts {
		addresses[contract.Address{
			Network: k,
			Address: v,
		}] = struct{}{}
	}
	return &TezosDomain{
		storage, addresses,
	}
}

// Do -
func (td *TezosDomain) Do(bmd *domains.BigMapDiff, storage *ast.TypedAst) ([]models.Model, error) {
	if len(td.contracts) == 0 {
		return nil, nil
	}

	if _, ok := td.contracts[contract.Address{
		Address: bmd.BigMap.Contract,
		Network: bmd.BigMap.Network,
	}]; !ok {
		return nil, nil
	}

	switch bmd.BigMap.Name {
	case recordsAnnot:
		items, err := td.updateRecordsTZIP(bmd)
		return items, err
	case expiryMapAnnot:
		items, err := td.updateExpirationDate(bmd)
		return items, err
	}
	return nil, nil
}

func (td *TezosDomain) updateRecordsTZIP(bmd *domains.BigMapDiff) ([]models.Model, error) {
	if len(bmd.KeyStrings) == 0 || len(bmd.ValueStrings) == 0 {
		return nil, errors.Errorf("Invalid tezos domains big map diff: %d", bmd.GetID())
	}
	address, err := td.getAddress(bmd.Value)
	if err != nil {
		return nil, err
	}
	tezosDomain := tezosdomain.TezosDomain{
		Network:   bmd.BigMap.Network,
		Name:      bmd.KeyStrings[0],
		Level:     bmd.Level,
		Timestamp: bmd.Timestamp,
	}
	if address != nil {
		tezosDomain.Address = *address
	}
	return []models.Model{&tezosDomain}, nil
}

func (td *TezosDomain) updateExpirationDate(bmd *domains.BigMapDiff) ([]models.Model, error) {
	if len(bmd.KeyStrings) == 0 {
		return nil, errors.Errorf("Invalid tezos domains big map diff: %d", bmd.GetID())
	}
	ts := gjson.ParseBytes(bmd.Value).Get("int").Int()
	date := time.Unix(ts, 0).UTC()
	tezosDomain := tezosdomain.TezosDomain{
		Name:       bmd.KeyStrings[0],
		Network:    bmd.BigMap.Network,
		Expiration: date,
	}
	return []models.Model{&tezosDomain}, nil
}

func (td *TezosDomain) getAddress(value []byte) (*string, error) {
	val := gjson.ParseBytes(value)
	s := val.Get("args.0.args.0.args.0.args.0.bytes").String()
	if s == "" {
		if val.Get("args.0.args.0.args.0.prim").String() == consts.None {
			return nil, nil
		}
		return nil, errors.Errorf("Can't parse tezos domain address in big map value: %s", value)
	}
	address, err := forge.UnforgeAddress(s)
	if err != nil {
		return nil, err
	}
	return &address, nil
}
