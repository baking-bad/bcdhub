package docstring

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/pkg/errors"
)

func getSimpleExpr(bPath string, md meta.Metadata) (string, error) {
	node := md[bPath]

	if isOption(bPath) {
		return fmt.Sprintf("option(%s)", node.Prim), nil
	}

	return node.Prim, nil
}

func getCompactExpr(dd *dsData, bPath string, md meta.Metadata) (string, error) {
	node := md[bPath]

	switch node.Prim {
	case consts.CONTRACT, consts.LAMBDA:
		varName, err := handleType(dd, bPath, md)
		if err != nil {
			return "", err
		}

		expr := fmt.Sprintf("%s(%s)", node.Prim, varName)

		if isOption(bPath) {
			return fmt.Sprintf("option(%s)", expr), nil
		}

		return expr, nil
	case consts.LIST, consts.SET:
		path := fmt.Sprintf("%s/%s", bPath, typePrefix[node.Prim])
		varName, err := getTypeExpr(dd, path, md)
		if err != nil {
			return "", err
		}

		expr := fmt.Sprintf("%s(%s)", node.Prim, varName)

		if isOption(bPath) {
			return fmt.Sprintf("option(%s)", expr), nil
		}

		return expr, nil
	case consts.OPTION:
		path := fmt.Sprintf("%s/%s", bPath, typePrefix[node.Prim])
		return getTypeExpr(dd, path, md)
	case consts.MAP, consts.BIGMAP:
		key := md[bPath+"/k"]
		val := md[bPath+"/v"]
		return fmt.Sprintf("%s(%s, %s)", node.Prim, key.Prim, val.Prim), nil
	case consts.OR, consts.PAIR:
		arg0 := md[node.Args[0]]
		arg1 := md[node.Args[1]]
		return fmt.Sprintf("%s(%s, %s)", node.Prim, arg0.Prim, arg1.Prim), nil
	default:
		return "", errors.Errorf("[getCompactExpr] unknown node type %##v %s", md[bPath], bPath)
	}
}

func getComplexExpr(dd *dsData, bPath string, md meta.Metadata) (string, error) {
	node := md[bPath]

	switch node.Prim {
	case consts.PAIR, consts.OR:
		varName, err := handleType(dd, bPath, md)
		if err != nil {
			return "", err
		}

		if isOption(bPath) {
			return fmt.Sprintf("option(%s)", varName), nil
		}

		return varName, nil
	case consts.MAP, consts.BIGMAP, consts.LAMBDA:
		return handleType(dd, bPath, md)
	default:
		return "", errors.Errorf("[getComplexExpr] unknown node type %##v %s", md[bPath], bPath)
	}
}

func isOption(path string) bool {
	return path[len(path)-1] == 'o'
}

func trimOption(path string) string {
	return strings.TrimSuffix(path, "/o")
}
