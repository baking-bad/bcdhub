package migrations

import (
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/schollz/progressbar/v3"
)

// SetOperationTags -
type SetOperationTags struct{}

// Key -
func (m *SetOperationTags) Key() string {
	return "set_operation_tags"
}

// Description -
func (m *SetOperationTags) Description() string {
	return "set operation tags (FA1.2 and FA2)"
}

// Do - migrate function
func (m *SetOperationTags) Do(ctx *config.Context) error {
	operations, err := ctx.Operations.Get(nil, 0, false)
	if err != nil {
		return err
	}
	logger.Info("Found %d operations", len(operations))

	result := make([]models.Model, 0)

	bar := progressbar.NewOptions(len(operations), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())

	tags := make(map[string][]string)
	for i := range operations {
		if err := bar.Add(1); err != nil {
			return err
		}

		if _, ok := tags[operations[i].Destination]; !ok {
			contract := contract.NewEmptyContract(operations[i].Network, operations[i].Destination)
			if err := ctx.Storage.GetByID(&contract); err != nil {
				if ctx.Storage.IsRecordNotFound(err) {
					continue
				}
				return err
			}
			operationTags := make([]string, 0)
			for _, tag := range contract.Tags {
				if helpers.StringInArray(tag, []string{
					consts.FA12Tag, consts.FA2Tag,
				}) {
					operationTags = append(operationTags, tag)
				}
			}
			tags[operations[i].Destination] = operationTags
		}
		operations[i].Tags = tags[operations[i].Destination]
		result = append(result, &operations[i])
	}

	if err := ctx.Storage.BulkUpdate(result); err != nil {
		logger.Errorf("ctx.Storage.BulkUpdate error: %v", err)
		return err
	}

	logger.Info("Done. %d operations were tagged.", len(result))

	return nil
}
