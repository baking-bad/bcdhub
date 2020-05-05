package docstring

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

func handleTupleEnumUnion(dd *dsData, bPath string, i int, md meta.Metadata) (string, error) {
	var node = md[bPath]
	var args []TypedefArg

	name, err := getVarName(dd, bPath, md)
	if err != nil {
		return "", err
	}

	for i, argPath := range node.Args {
		dd.arg = i
		value, err := getTypeExpr(dd, argPath, md)
		if err != nil {
			return "", err
		}
		args = append(args, TypedefArg{Value: value})
	}

	dd.insertTypedef(i, Typedef{
		Name: name,
		Type: node.Prim,
		Args: args,
	})

	return fmt.Sprintf("$%s", name), nil
}

func handleNamed(dd *dsData, bPath string, i int, md meta.Metadata) (string, error) {
	var node = md[bPath]
	var args []TypedefArg

	name, err := getVarName(dd, bPath, md)
	if err != nil {
		return "", err
	}

	for i, argPath := range node.Args {
		dd.arg = i
		value, err := getTypeExpr(dd, argPath, md)
		if err != nil {
			return "", err
		}

		args = append(args, TypedefArg{Key: md[argPath].Name, Value: value})
	}

	dd.insertTypedef(i, Typedef{
		Name: name,
		Type: node.Prim,
		Args: args,
	})

	return fmt.Sprintf("$%s", name), nil
}

func handleContractLambda(dd *dsData, bPath string, i int, md meta.Metadata) (string, error) {
	node := md[bPath]
	parsed := gjson.Parse(node.Parameter)
	parameter, err := formatter.MichelineToMichelson(parsed, true, formatter.DefLineSize)
	if err != nil {
		return "", err
	}

	if isSimpleParam(parameter) {
		return parameter, nil
	}

	name, err := getVarNameContractLambda(dd, bPath, md)
	if err != nil {
		return "", err
	}

	dd.insertTypedef(i, Typedef{
		Name: name,
		Type: parameter,
	})

	return fmt.Sprintf("$%s", name), nil
}

func handleMap(dd *dsData, bPath string, i int, md meta.Metadata) (string, error) {
	node := md[bPath]
	var args []TypedefArg
	name, err := getVarName(dd, bPath, md)
	if err != nil {
		return "", err
	}
	key, err := getTypeExpr(dd, bPath+"/k", md)
	if err != nil {
		return "", err
	}
	value, err := getTypeExpr(dd, bPath+"/v", md)
	if err != nil {
		return "", err
	}

	args = append(args, TypedefArg{Key: "key", Value: key})
	args = append(args, TypedefArg{Key: "value", Value: value})

	dd.insertTypedef(i, Typedef{
		Name: name,
		Type: node.Prim,
		Args: args,
	})

	return fmt.Sprintf("$%s", name), nil
}
