package handlers

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
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
	storage    models.GeneralRepository
	operations operation.Repository

	contracts map[contract.Address]struct{}
	ptrs      map[contract.Address]ptrs
}

type ptrs struct {
	records *int64
	expiry  *int64
}

// NewTezosDomains -
func NewTezosDomains(storage models.GeneralRepository, operations operation.Repository, contracts map[types.Network]string) *TezosDomain {
	addresses := make(map[contract.Address]struct{})
	for k, v := range contracts {
		addresses[contract.Address{
			Network: k,
			Address: v,
		}] = struct{}{}
	}
	return &TezosDomain{
		storage, operations, addresses, make(map[contract.Address]ptrs),
	}
}

// Do -
func (td *TezosDomain) Do(bmd *bigmapdiff.BigMapDiff, storage *ast.TypedAst) (bool, []models.Model, error) {
	bmd, ptr := td.getBigMapDiff(bmd, storage)
	if bmd == nil {
		return false, nil, nil
	}
	switch bmd.Ptr {
	case *ptr.records:
		items, err := td.updateRecordsTZIP(bmd)
		return true, items, err
	case *ptr.expiry:
		items, err := td.updateExpirationDate(bmd)
		return true, items, err
	}
	return false, nil, nil
}

func (td *TezosDomain) getBigMapDiff(bmd *bigmapdiff.BigMapDiff, storage *ast.TypedAst) (*bigmapdiff.BigMapDiff, *ptrs) {
	if len(td.contracts) == 0 {
		return nil, nil
	}
	address := contract.Address{
		Address: bmd.Contract,
		Network: bmd.Network,
	}
	if _, ok := td.contracts[address]; !ok {
		return nil, nil
	}
	ptr, ok := td.ptrs[address]
	if !ok {
		if err := td.getPointers(address, bmd, storage); err != nil {
			return nil, nil
		}
		ptr = td.ptrs[address]
	}

	return bmd, &ptr
}

func (td *TezosDomain) getPointers(address contract.Address, bmd *bigmapdiff.BigMapDiff, storage *ast.TypedAst) error {
	var res ptrs

	op, err := td.operations.GetByIDs(bmd.OperationID)
	if err != nil {
		if td.storage.IsRecordNotFound(err) {
			return nil
		}
		return err
	}
	if len(op) != 1 {
		return nil
	}

	var storageData ast.UntypedAST
	if err := json.Unmarshal(op[0].DeffatedStorage, &storageData); err != nil {
		return err
	}

	storageTree := ast.TypedAst{
		Nodes: []ast.Node{ast.Copy(storage.Nodes[0])},
	}

	if err := storageTree.Settle(storageData); err != nil {
		return err
	}

	for _, annot := range []string{recordsAnnot, expiryMapAnnot} {
		if node := storageTree.FindByName(annot, false); node != nil {
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

func (td *TezosDomain) updateRecordsTZIP(bmd *bigmapdiff.BigMapDiff) ([]models.Model, error) {
	if len(bmd.KeyStrings) == 0 || len(bmd.ValueStrings) == 0 {
		return nil, errors.Errorf("Invalid tezos domains big map diff: %d", bmd.GetID())
	}
	address, err := td.getAddress(bmd.Value)
	if err != nil {
		return nil, err
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
	return []models.Model{&tezosDomain}, nil
}

func (td *TezosDomain) updateExpirationDate(bmd *bigmapdiff.BigMapDiff) ([]models.Model, error) {
	if len(bmd.KeyStrings) == 0 {
		return nil, errors.Errorf("Invalid tezos domains big map diff: %d", bmd.GetID())
	}
	ts := gjson.ParseBytes(bmd.Value).Get("int").Int()
	date := time.Unix(ts, 0).UTC()
	tezosDomain := tezosdomain.TezosDomain{
		Name:       bmd.KeyStrings[0],
		Network:    bmd.Network,
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
