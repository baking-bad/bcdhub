package contractparser

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/node"
)

func primTags(node node.Node) string {
	switch node.Prim {
	case consts.CREATECONTRACT:
		return consts.ContractFactoryTag
	case consts.SETDELEGATE:
		return consts.DelegatableTag
	case consts.CHECKSIGNATURE:
		return consts.CheckSigTag
	case consts.CHAINID:
		return consts.ChainAwareTag
	}
	return ""
}
