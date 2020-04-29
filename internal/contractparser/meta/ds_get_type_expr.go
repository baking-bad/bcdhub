package meta

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
)

func (md Metadata) getSimpleExpr(bPath string) (string, error) {
	node := md[bPath]

	if isOption(bPath) {
		return fmt.Sprintf("option(%s)", node.Prim), nil
	}

	return node.Prim, nil
}

func (md Metadata) getCompactExpr(dd *dsData, bPath string) (string, error) {
	node := md[bPath]

	switch node.Prim {
	case consts.CONTRACT, consts.LAMBDA:
		varName, err := md.handleType(dd, bPath)
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
		varName, err := md.getTypeExpr(dd, path)
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
		return md.getTypeExpr(dd, path)
	case consts.MAP, consts.BIGMAP:
		key := md[bPath+"/k"]
		val := md[bPath+"/v"]
		return fmt.Sprintf("%s(%s, %s)", node.Prim, key.Prim, val.Prim), nil
	case consts.OR, consts.PAIR:
		arg0 := md[node.Args[0]]
		arg1 := md[node.Args[1]]
		return fmt.Sprintf("%s(%s, %s)", node.Prim, arg0.Prim, arg1.Prim), nil
	default:
		return "", fmt.Errorf("[getCompactExpr] unknown node type %##v %s", md[bPath], bPath)
	}
}

func (md Metadata) getComplexExpr(dd *dsData, bPath string) (string, error) {
	node := md[bPath]

	switch node.Prim {
	case consts.PAIR, consts.OR:
		varName, err := md.handleType(dd, bPath)
		if err != nil {
			return "", err
		}

		if isOption(bPath) {
			return fmt.Sprintf("option(%s)", varName), nil
		}

		return varName, nil
	case consts.MAP, consts.BIGMAP:
		return md.handleType(dd, bPath)
	default:
		return "", fmt.Errorf("[getComplexExpr] unknown node type %##v %s", md[bPath], bPath)
	}
}

func isOption(path string) bool {
	return path[len(path)-1] == 'o'
}
