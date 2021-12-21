package indexer

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
)

func createProtocol(rpc noderpc.INode, network types.Network, hash string, level int64) (protocol protocol.Protocol, err error) {
	logger.Info().Str("network", network.String()).Msgf("Creating new protocol %s starting at %d", hash, level)
	protocol.SymLink, err = bcd.GetProtoSymLink(hash)
	if err != nil {
		return
	}

	protocol.Alias = hash[:8]
	protocol.Network = network
	protocol.Hash = hash
	protocol.StartLevel = level

	err = setProtocolConstants(rpc, &protocol)

	return
}

func setProtocolConstants(rpc noderpc.INode, proto *protocol.Protocol) error {
	if proto.StartLevel > 0 {
		resp, err := rpc.GetNetworkConstants(proto.StartLevel)
		if err != nil {
			return err
		}
		proto.Constants = new(protocol.Constants)
		proto.Constants.CostPerByte = resp.CostPerByte
		proto.Constants.HardGasLimitPerOperation = resp.HardGasLimitPerOperation
		proto.Constants.HardStorageLimitPerOperation = resp.HardStorageLimitPerOperation
		proto.Constants.TimeBetweenBlocks = resp.BlockDelay()
	}

	return nil
}

func (bi *BoostIndexer) fetchExternalProtocols(ctx context.Context) error {
	logger.Info().Str("network", bi.Network.String()).Msg("Fetching external protocols")
	existingProtocols, err := bi.Protocols.GetByNetworkWithSort(bi.Network, "start_level", "desc")
	if err != nil {
		return err
	}

	exists := make(map[string]bool)
	for _, existingProtocol := range existingProtocols {
		exists[existingProtocol.Hash] = true
	}

	extProtocols, err := bi.externalIndexer.GetProtocols()
	if err != nil {
		return err
	}

	protocols := make([]models.Model, 0)
	for i := range extProtocols {
		if _, ok := exists[extProtocols[i].Hash]; ok {
			continue
		}
		symLink, err := bcd.GetProtoSymLink(extProtocols[i].Hash)
		if err != nil {
			return err
		}
		alias := extProtocols[i].Alias
		if alias == "" {
			alias = extProtocols[i].Hash[:8]
		}

		newProtocol := &protocol.Protocol{
			Hash:       extProtocols[i].Hash,
			Alias:      alias,
			StartLevel: extProtocols[i].StartLevel,
			EndLevel:   extProtocols[i].LastLevel,
			SymLink:    symLink,
			Network:    bi.Network,
			Constants: &protocol.Constants{
				CostPerByte:                  extProtocols[i].Constants.CostPerByte,
				HardStorageLimitPerOperation: extProtocols[i].Constants.HardStorageLimitPerOperation,
				HardGasLimitPerOperation:     extProtocols[i].Constants.HardGasLimitPerOperation,
				TimeBetweenBlocks:            extProtocols[i].Constants.TimeBetweenBlocks,
			},
		}

		protocols = append(protocols, newProtocol)
		logger.Info().Str("network", bi.Network.String()).Msgf("Fetched %s", alias)
	}

	return bi.Storage.Save(ctx, protocols)
}
