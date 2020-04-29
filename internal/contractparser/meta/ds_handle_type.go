package meta

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/tidwall/gjson"
)

func (md Metadata) handleTupleEnumUnion(dd *dsData, bPath string, i int) (string, error) {
	var node = md[bPath]
	var args []TypedefArg

	name, err := md.getVarName(dd, bPath)
	if err != nil {
		return "", err
	}

	for i, argPath := range node.Args {
		dd.arg = i
		value, err := md.getTypeExpr(dd, argPath)
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

func (md Metadata) handleNamed(dd *dsData, bPath string, i int) (string, error) {
	var node = md[bPath]
	var args []TypedefArg

	name, err := md.getVarName(dd, bPath)
	if err != nil {
		return "", err
	}

	for i, argPath := range node.Args {
		dd.arg = i
		value, err := md.getTypeExpr(dd, argPath)
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

func (md Metadata) handleContractLambda(dd *dsData, bPath string, i int) (string, error) {
	node := md[bPath]
	parsed := gjson.Parse(node.Parameter)
	parameter, err := formatter.MichelineToMichelson(parsed, true, formatter.DefLineSize)
	if err != nil {
		return "", err
	}

	if isSimpleParam(parameter) {
		return parameter, nil
	}

	name, err := md.getVarNameContractLambda(dd, bPath)
	if err != nil {
		return "", err
	}

	dd.insertTypedef(i, Typedef{
		Name: name,
		Type: parameter,
	})

	return fmt.Sprintf("$%s", name), nil
}

func (md Metadata) handleMap(dd *dsData, bPath string, i int) (string, error) {
	node := md[bPath]
	var args []TypedefArg
	name, err := md.getVarName(dd, bPath)
	if err != nil {
		return "", err
	}
	key, err := md.getTypeExpr(dd, bPath+"/k")
	if err != nil {
		return "", err
	}
	value, err := md.getTypeExpr(dd, bPath+"/v")
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
