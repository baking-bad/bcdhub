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
	"github.com/baking-bad/bcdhub/internal/models/operation"
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
	storage    models.GeneralRepository
	operations operation.Repository
	shareDir   string

	contracts map[contract.Address]struct{}
	ptrs      map[contract.Address]ptrs
}

type ptrs struct {
	records *int64
	expiry  *int64
}

// NewTezosDomains -
func NewTezosDomains(storage models.GeneralRepository, operations operation.Repository, contracts map[string]string, shareDir string) *TezosDomain {
	addresses := make(map[contract.Address]struct{})
	for k, v := range contracts {
		addresses[contract.Address{
			Network: k,
			Address: v,
		}] = struct{}{}
	}
	return &TezosDomain{
		storage, operations, shareDir, addresses, make(map[contract.Address]ptrs),
	}
}

// Do -
func (td *TezosDomain) Do(bmd *bigmapdiff.BigMapDiff) (bool, []models.Model, error) {
	bmd, ptr := td.getBigMapDiff(bmd)
	if bmd == nil {
		return false, nil, nil
	}
	switch bmd.Ptr {
	case *ptr.records:
		return true, nil, td.updateRecordsTZIP(bmd)
	case *ptr.expiry:
		return true, nil, td.updateExpirationDate(bmd)
	}
	return false, nil, nil
}

func (td *TezosDomain) getBigMapDiff(bmd *bigmapdiff.BigMapDiff) (*bigmapdiff.BigMapDiff, *ptrs) {
	if len(td.contracts) == 0 {
		return nil, nil
	}
	address := contract.Address{
		Address: bmd.Address,
		Network: bmd.Network,
	}
	if _, ok := td.contracts[address]; !ok {
		return nil, nil
	}
	ptr, ok := td.ptrs[address]
	if !ok {
		if err := td.getPointers(address, bmd.Protocol, bmd); err != nil {
			return nil, nil
		}
		ptr = td.ptrs[address]
	}

	return bmd, &ptr
}

func (td *TezosDomain) getPointers(address contract.Address, protocol string, bmd *bigmapdiff.BigMapDiff) error {
	var res ptrs
	data, err := fetch.Contract(address.Address, address.Network, protocol, td.shareDir)
	if err != nil {
		return err
	}

	tree, err := ast.NewTypedAstFromString(gjson.ParseBytes(data).Get("#(prim=\"storage\").args.0").Raw)
	if err != nil {
		return err
	}

	op, err := td.operations.GetOne(bmd.OperationHash, bmd.OperationCounter, bmd.OperationNonce)
	if err != nil {
		if td.storage.IsRecordNotFound(err) {
			return nil
		}
		return err
	}

	var storageData ast.UntypedAST
	if err := json.Unmarshal(op.DeffatedStorage, &storageData); err != nil {
		return err
	}
	if err := tree.Settle(storageData); err != nil {
		return err
	}

	for _, annot := range []string{recordsAnnot, expiryMapAnnot} {
		if node := tree.FindByName(annot, false); node != nil {
			if b, ok := node.(*ast.BigMap); ok && b.Ptr != nil {
				switch annot {
				case recordsAnnot:
					res.records = b.Ptr
				case expiryMapAnnot:
					res.expiry = b.Ptr
				}
			}
		}
	}

	td.ptrs[address] = res
	return nil
}

func (td *TezosDomain) updateRecordsTZIP(bmd *bigmapdiff.BigMapDiff) error {
	if len(bmd.KeyStrings) == 0 || len(bmd.ValueStrings) == 0 {
		return errors.Errorf("Invalid tezos domains big map diff: %d", bmd.GetID())
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
		return errors.Errorf("Invalid tezos domains big map diff: %d", bmd.GetID())
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
