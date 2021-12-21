package services

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/types"
	tzipStorage "github.com/baking-bad/bcdhub/internal/parsers/tzip/storage"
	"github.com/baking-bad/bcdhub/internal/parsers/tzip/tokens"
	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"
)

// Unknown -
type Unknown struct {
	*TimeBased
	ctx     *config.Context
	timeout time.Duration
	since   time.Duration
}

// NewUnknown -
func NewUnknown(ctx *config.Context, period, timeout, since time.Duration) *Unknown {
	u := &Unknown{
		ctx:     ctx,
		timeout: timeout,
		since:   since,
	}
	u.TimeBased = NewTimeBased(u.refresh, period)
	return u
}

func (u *Unknown) refresh(ctx context.Context) error {
	since := time.Now().Add(u.since)
	metadata, err := u.ctx.TokenMetadata.GetRecent(since, tokenmetadata.GetContext{
		Name: consts.Unknown,
	})
	if err != nil {
		return err
	}
	logger.Info().Msgf("Found %d unknown token metadata", len(metadata))

	ipfs := tzipStorage.NewIPFSStorage(u.ctx.Config.IPFSGateways, tzipStorage.WithTimeoutIPFS(u.timeout))

	return u.ctx.StorageDB.DB.RunInTransaction(ctx, func(tx *pg.Tx) error {
		for i := range metadata {
			emptyValue, ok := metadata[i].Extras["@@empty"]
			if !ok {
				continue
			}
			link, ok := emptyValue.(string)
			if !ok {
				continue
			}

			remoteMetadata := new(tokens.TokenMetadata)
			if err := ipfs.Get(link, remoteMetadata); err != nil {
				if errors.Is(err, tzipStorage.ErrNoIPFSResponse) || errors.Is(err, tzipStorage.ErrInvalidIPFSHash) {
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
			logger.Info().Str("url", link).Msg("token metadata fetched")
		}

		return nil
	})
}
