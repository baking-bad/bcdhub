package formatter

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/baking-bad/bcdhub/internal/helpers"
)

// DefLineSize -
const DefLineSize = 88

// IsFramed -
func IsFramed(n gjson.Result) bool {
	prim := n.Get("prim").String()
	if helpers.StringInArray(prim, []string{
		"Pair", "Left", "Right", "Some",
		"pair", "or", "option", "map", "big_map", "list", "set", "contract", "lambda",
		"ticket", "sapling_state", "sapling_transaction",
	}) {
		return true
	} else if helpers.StringInArray(prim, []string{
		"key", "unit", "signature", "operation",
		"int", "nat", "string", "bytes", "mutez", "bool", "key_hash", "timestamp", "address",
		"bls12_381_g1", "bls12_381_g2", "bls12_381_fr", "chain_id", "never",
	}) {
		return n.Get("annots").Exists()
	}
	return false
}

// IsComplex -
func IsComplex(n gjson.Result) bool {
	prim := n.Get("prim").String()
	return prim == "LAMBDA" || prim[:2] == "IF"
}

// IsInline -
func IsInline(n gjson.Result) bool {
	prim := n.Get("prim").String()
	return prim == "PUSH"
}

// IsScript -
func IsScript(n gjson.Result) bool {
	if !n.IsArray() {
		return false
	}
	for _, item := range n.Array() {
		prim := item.Get("prim").String()
		if !helpers.StringInArray(prim, []string{
			"parameter", "storage", "code",
		}) {
			return false
		}
	}
	return true
}

// MichelineStringToMichelson -
func MichelineStringToMichelson(str string, inline bool, lineSize int) (string, error) {
	return MichelineToMichelson(gjson.Parse(str), inline, lineSize)
}

// MichelineToMichelson -
func MichelineToMichelson(n gjson.Result, inline bool, lineSize int) (string, error) {
	return formatNode(n, "", inline, true, false, lineSize)
}

func formatNode(node gjson.Result, indent string, inline, isRoot, wrapped bool, lineSize int) (string, error) {
	if node.IsArray() {
		return formatArray(node, indent, inline, isRoot, lineSize)
	}

	if node.IsObject() {
		return formatObject(node, indent, inline, isRoot, wrapped, lineSize)
	}

	return "", errors.Errorf("data is not array or object %v", node)
}

func formatArray(node gjson.Result, indent string, inline, isRoot bool, lineSize int) (string, error) {
	seqIndent := indent
	isScriptRoot := isRoot && IsScript(node)
	if !isScriptRoot {
		seqIndent = indent + "  "
	}

	items := make([]string, len(node.Array()))

	for i, n := range node.Array() {
		res, err := formatNode(n, seqIndent, inline, false, true, lineSize)
		if err != nil {
			return "", err
		}
		items[i] = res
	}

	if len(items) == 0 {
		return "{}", nil
	}

	length := len(indent) + 4
	for _, i := range items {
		length += len(i)
	}

	space := ""
	if !isScriptRoot {
		space = " "
	}

	var seq string

	if inline || length < lineSize {
		seq = strings.Join(items, fmt.Sprintf("%v; ", space))
	} else {
		seq = strings.Join(items, fmt.Sprintf("%v;\n%v", space, seqIndent))
	}

	if !isScriptRoot {
		return fmt.Sprintf("{ %v }", seq), nil
	}

	return seq, nil
}

func formatObject(node gjson.Result, indent string, inline, isRoot, wrapped bool, lineSize int) (string, error) {
	if node.Get("prim").Exists() {
		return formatPrimObject(node, indent, inline, isRoot, wrapped, lineSize)
	}

	return formatNonPrimObject(node)
}

func formatPrimObject(node gjson.Result, indent string, inline, isRoot, wrapped bool, lineSize int) (string, error) {
	res := []string{node.Get("prim").String()}

	if annots := node.Get("annots"); annots.Exists() {
		for _, a := range annots.Array() {
			res = append(res, a.String())
		}
	}

	expr := strings.Join(res, " ")

	var args []gjson.Result
	if rawArgs := node.Get("args"); rawArgs.Exists() {
		args = rawArgs.Array()
	}

	switch {
	case IsComplex(node):
		argIndent := indent + "  "
		items := make([]string, len(args))
		for i, a := range args {
			res, err := formatNode(a, argIndent, inline, false, false, lineSize)
			if err != nil {
				return "", err
			}

			items[i] = res
		}

		length := len(indent) + len(expr) + len(items) + 1

		for _, item := range items {
			length += len(item)
		}

		if inline || length < lineSize {
			expr = fmt.Sprintf("%v %v", expr, strings.Join(items, " "))
		} else {
			res := []string{expr}
			res = append(res, items...)
			expr = strings.Join(res, fmt.Sprintf("\n%v", argIndent))
		}
	case len(args) == 1:
		argIndent := indent + strings.Repeat(" ", len(expr)+1)
		res, err := formatNode(args[0], argIndent, inline, false, false, lineSize)
		if err != nil {
			return "", err
		}
		expr = fmt.Sprintf("%v %v", expr, res)
	case len(args) > 1:
		argIndent := indent + "  "
		altIndent := indent + strings.Repeat(" ", len(expr)+2)

		for _, arg := range args {
			item, err := formatNode(arg, argIndent, inline, false, false, lineSize)
			if err != nil {
				return "", err
			}
			length := len(indent) + len(expr) + len(item) + 1
			if inline || IsInline(node) || length < lineSize {
				argIndent = altIndent
				expr = fmt.Sprintf("%v %v", expr, item)
			} else {
				expr = fmt.Sprintf("%v\n%v%v", expr, argIndent, item)
			}
		}
	}

	if IsFramed(node) && !isRoot && !wrapped {
		return fmt.Sprintf("(%v)", expr), nil
	}
	return expr, nil
}

func formatNonPrimObject(node gjson.Result) (string, error) {
	if len(node.Map()) != 1 {
		return "", errors.Errorf("node keys count != 1: %v", node)
	}

	for coreType, value := range node.Map() {
		switch coreType {
		case "int":
			return value.String(), nil
		case "bytes":
			return fmt.Sprintf("0x%v", value.String()), nil
		case "string":
			return value.Raw, nil
		}
	}

	return "", errors.Errorf("invalid coreType: %v", node)
}
