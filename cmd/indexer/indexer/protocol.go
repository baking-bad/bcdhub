package indexer

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/noderpc"
)

func createProtocol(ctx context.Context, rpc noderpc.INode, hash string, level int64) (protocol protocol.Protocol, err error) {
	protocol.SymLink, err = bcd.GetProtoSymLink(hash)
	if err != nil {
		return
	}

	protocol.Alias = hash[:8]
	protocol.Hash = hash
	protocol.StartLevel = level

	err = setProtocolConstants(ctx, rpc, &protocol)

	return
}

func setProtocolConstants(ctx context.Context, rpc noderpc.INode, proto *protocol.Protocol) error {
	if proto.StartLevel > 0 {
		resp, err := rpc.GetNetworkConstants(ctx, proto.StartLevel)
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
