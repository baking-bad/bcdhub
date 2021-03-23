package operations

import (
	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
)

func setTags(contractRepo contract.Repository, generalRepo models.GeneralRepository, op *operation.Operation) error {
	if !bcd.IsContract(op.Destination) {
		return nil
	}

	c, err := contractRepo.Get(op.Network, op.Destination)
	if err != nil {
		if generalRepo.IsRecordNotFound(err) {
			return nil
		}
		return err
	}
	for _, tag := range c.Tags {
		switch tag {
		case consts.FA12Tag, consts.FA2Tag:
			if op.Tags == nil {
				op.Tags = make([]string, 0)
			}
			op.Tags = append(op.Tags, tag)
		case consts.NFTLedgerTag:
			if op.Tags == nil {
				op.Tags = make([]string, 0)
			}
			op.Tags = append(op.Tags, tag)
		}
	}
	return nil
}
