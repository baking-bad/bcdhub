package indexer

import (
	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/noderpc"
)

func createProtocol(rpc noderpc.INode, network, hash string, level int64) (protocol protocol.Protocol, err error) {
	logger.WithNetwork(network).Infof("Creating new protocol %s starting at %d", hash, level)
	protocol.SymLink, err = bcd.GetProtoSymLink(hash)
	if err != nil {
		return
	}

	protocol.Alias = hash[:8]
	protocol.Network = network
	protocol.Hash = hash
	protocol.StartLevel = level
	protocol.ID = helpers.GenerateID()

	err = setProtocolConstants(rpc, &protocol)

	return
}

func setProtocolConstants(rpc noderpc.INode, proto *protocol.Protocol) error {
	if proto.StartLevel > 0 {
		resp, err := rpc.GetNetworkConstants(proto.StartLevel)
		if err != nil {
			return err
		}

		proto.Constants.CostPerByte = resp.CostPerByte
		proto.Constants.HardGasLimitPerOperation = resp.HardGasLimitPerOperation
		proto.Constants.HardStorageLimitPerOperation = resp.HardStorageLimitPerOperation
		proto.Constants.TimeBetweenBlocks = resp.TimeBetweenBlocks[0]
	}

	return nil
}
