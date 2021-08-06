package services

import (
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
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Unknown -
type Unknown struct {
	*TimeBased
	ctx *config.Context
}

// NewUnknown -
func NewUnknown(ctx *config.Context, period time.Duration) *Unknown {
	u := &Unknown{
		ctx: ctx,
	}
	u.TimeBased = NewTimeBased(u.refresh, period)
	return u
}

func (u *Unknown) refresh() error {
	metadata, err := u.ctx.TokenMetadata.GetAll(tokenmetadata.GetContext{
		Name: consts.Unknown,
	})
	if err != nil {
		return err
	}
	logger.Info().Msgf("Found %d unknown metadata", len(metadata))

	return u.ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		for i := range metadata {
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

			s := tzipStorage.NewIPFSStorage(u.ctx.Config.IPFSGateways, tzipStorage.WithTimeoutIPFS(time.Second*10))

			remoteMetadata := new(tokens.TokenMetadata)
			if err := s.Get(link, remoteMetadata); err != nil {
				if errors.Is(err, tzipStorage.ErrNoIPFSResponse) {
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

			if err := metadata[i].Save(tx); err != nil {
				return err
			}
		}

		return nil
	})
}
