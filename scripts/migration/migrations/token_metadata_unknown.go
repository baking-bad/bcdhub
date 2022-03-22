package migrations

import (
	"context"
	"errors"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/types"
	tzipStorage "github.com/baking-bad/bcdhub/internal/parsers/contract_metadata/storage"
	"github.com/baking-bad/bcdhub/internal/parsers/contract_metadata/tokens"
	"github.com/go-pg/pg/v10"
	"github.com/schollz/progressbar/v3"
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
	metadata, err := ctx.TokenMetadata.GetRecent(time.Time{}, tokenmetadata.GetContext{
		Name: consts.Unknown,
	})
	if err != nil {
		return err
	}
	logger.Info().Msgf("Found %d unknown metadata", len(metadata))

	ipfs := tzipStorage.NewIPFSStorage(ctx.Config.IPFSGateways, tzipStorage.WithTimeoutIPFS(time.Second*5))
	bar := progressbar.NewOptions(len(metadata), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())

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

		remoteMetadata := new(tokens.TokenMetadata)
		if err := ipfs.Get(context.Background(), link, remoteMetadata); err != nil {
			if errors.Is(err, tzipStorage.ErrNoIPFSResponse) || errors.Is(err, tzipStorage.ErrInvalidIPFSHash) {
				logger.Warning().Err(err).Str("url", link).Str("kind", "token_metadata").Msg("")
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

		err := ctx.StorageDB.DB.RunInTransaction(context.Background(), func(tx *pg.Tx) error {
			return metadata[i].Save(tx)
		})
		if err != nil {
			return err
		}
	}

	return nil
}
