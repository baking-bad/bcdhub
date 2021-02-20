package handlers

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/fetch"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

const (
	recordsAnnot   = "records"
	expiryMapAnnot = "expiry_map"
)

// TezosDomain -
type TezosDomain struct {
	storage  models.GeneralRepository
	shareDir string

	contracts map[contract.Address]struct{}
	trees     map[contract.Address]*ast.TypedAst
}

// NewTezosDomains -
func NewTezosDomains(storage models.GeneralRepository, contracts map[string]string, shareDir string) *TezosDomain {
	addresses := make(map[contract.Address]struct{})
	for k, v := range contracts {
		addresses[contract.Address{
			Network: k,
			Address: v,
		}] = struct{}{}
	}
	return &TezosDomain{
		storage, shareDir, addresses, make(map[contract.Address]*ast.TypedAst),
	}
}

// Do -
func (td *TezosDomain) Do(model models.Model) (bool, error) {
	bmd, handler := td.getBigMapDiff(model)
	if bmd == nil {
		return false, nil
	}
	switch handler {
	case recordsAnnot:
		return true, td.updateRecordsTZIP(bmd)
	case expiryMapAnnot:
		return true, td.updateExpirationDate(bmd)
	}
	return false, nil
}

func (td *TezosDomain) getBigMapDiff(model models.Model) (*bigmapdiff.BigMapDiff, string) {
	if len(td.contracts) == 0 {
		return nil, ""
	}
	bmd, ok := model.(*bigmapdiff.BigMapDiff)
	if !ok {
		return nil, ""
	}
	address := contract.Address{
		Address: bmd.Address,
		Network: bmd.Network,
	}
	if _, ok := td.contracts[address]; !ok {
		return nil, ""
	}
	tree, err := td.getTree(address, bmd.Protocol)
	if err != nil {
		return nil, ""
	}

	if node := tree.FindByName(recordsAnnot); node != nil {
		bigMap, ok := node.(*ast.BigMap)
		if ok && *bigMap.Ptr == bmd.Ptr {
			return bmd, recordsAnnot
		}
	}

	if node := tree.FindByName(expiryMapAnnot); node != nil {
		bigMap, ok := node.(*ast.BigMap)
		if ok && *bigMap.Ptr == bmd.Ptr {
			return bmd, expiryMapAnnot
		}
	}
	return nil, ""
}

func (td *TezosDomain) getTree(address contract.Address, protocol string) (*ast.TypedAst, error) {
	tree, ok := td.trees[address]
	if ok {
		return tree, nil
	}
	data, err := fetch.Contract(address.Address, address.Network, protocol, td.shareDir)
	if err != nil {
		return nil, err
	}
	tree, err = ast.NewTypedAstFromBytes(data)
	if err != nil {
		return nil, err
	}
	td.trees[address] = tree
	return tree, nil
}

func (td *TezosDomain) updateRecordsTZIP(bmd *bigmapdiff.BigMapDiff) error {
	if len(bmd.KeyStrings) == 0 || len(bmd.ValueStrings) == 0 {
		return errors.Errorf("Invalid tezos domains big map diff: %s", bmd.GetID())
	}
	address, err := td.getAddress(bmd.Value)
	if err != nil {
		return err
	}
	tezosDomain := tezosdomain.TezosDomain{
		Network:   bmd.Network,
		Name:      bmd.KeyStrings[0],
		Level:     bmd.Level,
		Timestamp: bmd.Timestamp,
	}
	if address != nil {
		tezosDomain.Address = *address
	}
	return td.storage.UpdateFields(models.DocTezosDomains, tezosDomain.GetID(), tezosDomain, "Name", "Address", "Network", "Level", "Timestamp")
}

func (td *TezosDomain) updateExpirationDate(bmd *bigmapdiff.BigMapDiff) error {
	if len(bmd.KeyStrings) == 0 {
		return errors.Errorf("Invalid tezos domains big map diff: %s", bmd.GetID())
	}
	ts := gjson.ParseBytes(bmd.Value).Get("int").Int()
	date := time.Unix(ts, 0).UTC()
	tezosDomain := tezosdomain.TezosDomain{
		Name:       bmd.KeyStrings[0],
		Network:    bmd.Network,
		Expiration: date,
	}
	return td.storage.UpdateFields(models.DocTezosDomains, tezosDomain.GetID(), tezosDomain, "Expiration")
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
