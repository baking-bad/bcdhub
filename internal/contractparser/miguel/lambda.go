package miguel

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type lambdaDecoder struct{}

// Decode -
func (l *lambdaDecoder) Decode(node gjson.Result, path string, nm *meta.NodeMetadata, metadata meta.Metadata, isRoot bool) (interface{}, error) {
	val, err := formatter.MichelineToMichelson(node, false, formatter.DefLineSize)
	return map[string]interface{}{
		"miguel_value": val,
		"miguel_type":  nm.Type,
	}, err
}
