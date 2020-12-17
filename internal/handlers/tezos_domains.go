package handlers

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/unpack"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
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
	es elastic.IElastic

	contracts map[elastic.Address]struct{}
	metadata  map[elastic.Address]meta.Metadata
}

// NewTezosDomains -
func NewTezosDomains(es elastic.IElastic, contracts map[string]string) *TezosDomain {
	addresses := make(map[elastic.Address]struct{})
	for k, v := range contracts {
		addresses[elastic.Address{
			Network: k,
			Address: v,
		}] = struct{}{}
	}
	return &TezosDomain{
		es, addresses, make(map[elastic.Address]meta.Metadata),
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
	address := elastic.Address{
		Address: bmd.Address,
		Network: bmd.Network,
	}
	if _, ok := td.contracts[address]; !ok {
		return nil, ""
	}
	metadata, err := td.getMetadata(address, bmd.Protocol)
	if err != nil {
		return nil, ""
	}

	binPath := metadata.Find(recordsAnnot)
	if binPath == bmd.BinPath {
		return bmd, recordsAnnot
	}

	binPath = metadata.Find(expiryMapAnnot)
	if binPath == bmd.BinPath {
		return bmd, expiryMapAnnot
	}
	return nil, ""
}

func (td *TezosDomain) getMetadata(address elastic.Address, protocol string) (meta.Metadata, error) {
	metadata, ok := td.metadata[address]
	if ok {
		return metadata, nil
	}
	metadata, err := meta.GetMetadata(td.es, address.Address, consts.STORAGE, protocol)
	if err != nil {
		return metadata, err
	}
	td.metadata[address] = metadata
	return metadata, nil
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
	return td.es.UpdateFields(elastic.DocTezosDomains, tezosDomain.GetID(), tezosDomain, "Name", "Address", "Network", "Level", "Timestamp")
}

func (td *TezosDomain) updateExpirationDate(bmd *bigmapdiff.BigMapDiff) error {
	if len(bmd.KeyStrings) == 0 {
		return errors.Errorf("Invalid tezos domains big map diff: %s", bmd.GetID())
	}
	ts := gjson.Parse(bmd.Value).Get("int").Int()
	date := time.Unix(ts, 0).UTC()
	tezosDomain := tezosdomain.TezosDomain{
		Name:       bmd.KeyStrings[0],
		Network:    bmd.Network,
		Expiration: date,
	}
	return td.es.UpdateFields(elastic.DocTezosDomains, tezosDomain.GetID(), tezosDomain, "Expiration")
}

func (td *TezosDomain) getAddress(value string) (*string, error) {
	val := gjson.Parse(value)
	s := val.Get("args.0.args.0.args.0.args.0.bytes").String()
	if s == "" {
		if val.Get("args.0.args.0.args.0.prim").String() == consts.None {
			return nil, nil
		}
		return nil, errors.Errorf("Can't parse tezos domain address in big map value: %s", value)
	}
	address, err := unpack.Address(s)
	if err != nil {
		return nil, err
	}
	return &address, nil
}
