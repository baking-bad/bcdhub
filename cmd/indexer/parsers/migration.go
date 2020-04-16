package parsers

import (
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

// MigrationParser -
type MigrationParser struct {
	rpc            noderpc.Pool
	es             *elastic.Elastic
	filesDirectory string
}

// NewMigrationParser -
func NewMigrationParser(rpc noderpc.Pool, es *elastic.Elastic, filesDirectory string) *MigrationParser {
	return &MigrationParser{
		rpc:            rpc,
		es:             es,
		filesDirectory: filesDirectory,
	}
}

// Parse -
func (p *MigrationParser) Parse(data gjson.Result, head noderpc.Header, old models.Contract, prevProtocol string) (*models.Migration, error) {
	protoSymLink, err := meta.GetProtoSymLink(head.Protocol)
	if err != nil {
		return nil, err
	}
	if err := updateMetadata(p.es, data, protoSymLink, &old); err != nil {
		return nil, err
	}

	newFingerprint, err := computeFingerprint(data)
	if err != nil {
		return nil, err
	}
	if newFingerprint.Compare(old.Fingerprint) {
		return nil, nil
	}

	op := models.Migration{
		ID:          strings.ReplaceAll(uuid.New().String(), "-", ""),
		IndexedTime: time.Now().UnixNano() / 1000,

		Network:      old.Network,
		Level:        head.Level - 1,
		Protocol:     head.Protocol,
		PrevProtocol: prevProtocol,
		Address:      old.Address,
		Timestamp:    head.Timestamp,
		Kind:         consts.MigrationUpdate,
	}
	if _, err := p.es.UpdateDoc(elastic.DocContracts, old.ID, old); err != nil {
		return nil, err
	}
	return &op, nil
}
