package contractparser

import (
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/node"
)

func detectLiquidity(n node.Node, entries []string) bool {
	for _, a := range n.Annotations {
		if strings.Contains(a, "_slash_") {
			return true
		}
	}

	for _, e := range entries {
		if strings.Contains(e, "_Liq_entry") {
			return true
		}
	}
	return false
}

func detectPython(n node.Node) bool {
	str := n.GetString()
	if str == "" {
		return false
	}

	return strings.Contains(str, "SmartPy") || strings.Contains(str, "self.") || strings.Contains(str, "sp.")
}

func detectLIGO(n node.Node) bool {
	return n.GetString() == "GET_FORCE"
}
