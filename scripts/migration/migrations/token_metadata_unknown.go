package migrations

import (
	"errors"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/types"
	tzipStorage "github.com/baking-bad/bcdhub/internal/parsers/tzip/storage"
	"github.com/baking-bad/bcdhub/internal/parsers/tzip/tokens"
	"github.com/schollz/progressbar/v3"
	"gorm.io/gorm"
)

// TokenMetadataUnknown - migration that requests again token metadata
type TokenMetadataUnknown struct{}

// Key -
func (m *TokenMetadataUnknown) Key() string {
	return "token_metadata_unknown"
}

// Description -
func (m *TokenMetadataUnknown) Description() string {
	return "migration that requests again token metadata"
}

// Do - migrate function
func (m *TokenMetadataUnknown) Do(ctx *config.Context) error {
	metadata, err := ctx.TokenMetadata.GetAll(tokenmetadata.GetContext{
		Name: consts.Unknown,
	})
	if err != nil {
		return err
	}
	logger.Info("Found %d unknown metadata", len(metadata))

	bar := progressbar.NewOptions(len(metadata), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())

	return ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		for i := range metadata {
			if err := bar.Add(1); err != nil {
				return err
			}

			emptyValue, ok := metadata[i].Extras["@@empty"]
			if !ok {
				continue
			}
			link, ok := emptyValue.(string)
			if !ok {
				continue
			}

			if !helpers.IsIPFS(strings.TrimPrefix(link, "ipfs://")) {
				continue
			}

			s := tzipStorage.NewIPFSStorage(ctx.Config.IPFSGateways, tzipStorage.WithTimeoutIPFS(time.Second*10))

			remoteMetadata := new(tokens.TokenMetadata)
			if err := s.Get(link, remoteMetadata); err != nil {
				if errors.Is(err, tzipStorage.ErrNoIPFSResponse) {
					logger.WithField("url", link).WithField("kind", "token_metadata").Warning(err)
					continue
				}
				return err
			}
			metadata[i].Symbol = remoteMetadata.Symbol
			metadata[i].Decimals = remoteMetadata.Decimals
			metadata[i].Name = remoteMetadata.Name
			metadata[i].Description = remoteMetadata.Description
			metadata[i].ArtifactURI = remoteMetadata.ArtifactURI
			metadata[i].DisplayURI = remoteMetadata.DisplayURI
			metadata[i].ThumbnailURI = remoteMetadata.ThumbnailURI
			metadata[i].ExternalURI = remoteMetadata.ExternalURI
			metadata[i].IsTransferable = remoteMetadata.IsTransferable
			metadata[i].IsBooleanAmount = remoteMetadata.IsBooleanAmount
			metadata[i].ShouldPreferSymbol = remoteMetadata.ShouldPreferSymbol
			metadata[i].Creators = remoteMetadata.Creators
			metadata[i].Tags = remoteMetadata.Tags
			metadata[i].Formats = types.Bytes(remoteMetadata.Formats)
			metadata[i].Extras = remoteMetadata.Extras

			if err := metadata[i].Save(tx); err != nil {
				return err
			}
		}

		return nil
	})
}
