package migrations

import (
	"encoding/json"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/pkg/errors"
	"github.com/schollz/progressbar/v3"
)

// NFTMetadtaa -
type NFTMetadata struct{}

// Key -
func (m *NFTMetadata) Key() string {
	return "nft_metadata"
}

// Description -
func (m *NFTMetadata) Description() string {
	return "get all NFT fields from extras and set it to models`s fields"
}

// Do - migrate function
func (m *NFTMetadata) Do(ctx *config.Context) error {
	logger.Info("Getting all token metadata...")

	if err := ctx.Storage.(*core.Postgres).DB.AutoMigrate(&tokenmetadata.TokenMetadata{}); err != nil {
		return err
	}

	metadata, err := ctx.TokenMetadata.GetWithExtras()
	if err != nil {
		return err
	}

	logger.Info("Found %d metadata with extra fields", len(metadata))

	updated := make([]models.Model, len(metadata))

	bar := progressbar.NewOptions(len(metadata), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
	for i := range metadata {
		if err := bar.Add(1); err != nil {
			return err
		}

		for key, value := range metadata[i].Extras {
			switch key {
			case "description":
				if s, ok := value.(string); ok {
					metadata[i].Description = s
					delete(metadata[i].Extras, key)
				}
			case "artifactUri":
				if s, ok := value.(string); ok {
					metadata[i].ArtifactURI = s
					delete(metadata[i].Extras, key)
				}
			case "displayUri":
				if s, ok := value.(string); ok {
					metadata[i].DisplayURI = s
					delete(metadata[i].Extras, key)
				}
			case "thumbnailUri":
				if s, ok := value.(string); ok {
					metadata[i].ThumbnailURI = s
					delete(metadata[i].Extras, key)
				}
			case "externalUri":
				if s, ok := value.(string); ok {
					metadata[i].ExternalURI = s
					delete(metadata[i].Extras, key)
				}
			case "isTransferable":
				if b, ok := value.(bool); ok {
					metadata[i].IsTransferable = b
					delete(metadata[i].Extras, key)
				}
			case "isBooleanAmount":
				if b, ok := value.(bool); ok {
					metadata[i].IsBooleanAmount = b
					delete(metadata[i].Extras, key)
				}
			case "shouldPreferSymbol":
				if b, ok := value.(bool); ok {
					metadata[i].ShouldPreferSymbol = b
					delete(metadata[i].Extras, key)
				}
			case "formats":
				switch val := value.(type) {
				case []interface{}:
					metadata[i].Formats, err = json.Marshal(val)
					if err == nil {
						continue
					}
					delete(metadata[i].Extras, key)
				case string:
					metadata[i].Formats = []byte(val)
					delete(metadata[i].Extras, key)
				default:
					return errors.Errorf("Unknown formats type: %T", value)
				}
			case "tags":
				if tags, ok := value.([]interface{}); ok {
					metadata[i].Tags = make([]string, 0)
					for _, tag := range tags {
						if t, ok := tag.(string); ok {
							metadata[i].Tags = append(metadata[i].Tags, t)
						}
					}
					delete(metadata[i].Extras, key)
				}
			case "creators":
				if creators, ok := value.([]interface{}); ok {
					metadata[i].Creators = make([]string, 0)
					for _, creator := range creators {
						if c, ok := creator.(string); ok {
							metadata[i].Creators = append(metadata[i].Creators, c)
						}
					}
					delete(metadata[i].Extras, key)
				}
			}
		}

		updated[i] = &metadata[i]
	}

	return ctx.Storage.BulkUpdate(updated)
}
