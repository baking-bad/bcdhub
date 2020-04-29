package meta

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
)

// DocEntrypoint -
type DocEntrypoint struct {
	Name string    `json:"name"`
	Type []Typedef `json:"typedef"`
}

// Typedef -
type Typedef struct {
	Name string       `json:"name"`
	Type string       `json:"type"`
	Args []TypedefArg `json:"args,omitempty"`
}

// TypedefArg -
type TypedefArg struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value"`
}

type dsData struct {
	typedef []Typedef
	counter int
	arg     int
}

// GetDocEntrypoints -
func (md Metadata) GetDocEntrypoints() ([]DocEntrypoint, error) {
	root := md["0"]
	entrypoints := make([]DocEntrypoint, 0)

	if isComplexRoot(root) {
		for i, binPath := range root.Args {
			md[binPath].Name = md[binPath].GetEntrypointName(i)

			typeDefs, err := md.parseEntrypointTypes(binPath)
			if err != nil {
				return nil, err
			}

			entrypoints = append(entrypoints, DocEntrypoint{
				Name: md[binPath].Name,
				Type: typeDefs,
			})
		}
	} else {
		root.Name = root.GetEntrypointName(-1)
		typeDefs, err := md.parseEntrypointTypes("0")
		if err != nil {
			return nil, err
		}

		entrypoints = append(entrypoints, DocEntrypoint{
			Name: root.Name,
			Type: typeDefs,
		})
	}

	return entrypoints, nil
}

func isComplexRoot(root *NodeMetadata) bool {
	return len(root.Args) > 0 && root.Prim == consts.OR && (root.Type == consts.TypeUnion || root.Type == consts.TypeNamedEnum || root.Type == consts.TypeNamedTuple || root.Type == consts.TypeNamedUnion)
}

func (md Metadata) parseEntrypointTypes(bPath string) ([]Typedef, error) {
	var dd dsData

	typ, err := md.getTypeExpr(&dd, bPath)
	if err != nil {
		return nil, err
	}

	node := md[bPath]

	if isSimpleType(node.Prim) || md.isCompactType(bPath) {
		dd.insertTypedef(0, Typedef{
			Name: node.Name,
			Type: typ,
		})
	}

	return dd.typedef, nil
}

func (md Metadata) handleType(dd *dsData, bPath string) (string, error) {
	i := len(dd.typedef)

	switch md[bPath].Type {
	case consts.TypeTuple, consts.TypeEnum, consts.TypeUnion:
		return md.handleTupleEnumUnion(dd, bPath, i)
	case consts.TypeNamedTuple, consts.TypeNamedEnum, consts.TypeNamedUnion:
		return md.handleNamed(dd, bPath, i)
	case consts.CONTRACT, consts.LAMBDA:
		return md.handleContractLambda(dd, bPath, i)
	case consts.MAP, consts.BIGMAP:
		return md.handleMap(dd, bPath, i)
	default:
		return "", fmt.Errorf("[handleType] %##v %s", md[bPath], bPath)
	}
}

func (md Metadata) getTypeExpr(dd *dsData, bPath string) (string, error) {
	nodeType, err := md.getType(bPath)
	if err != nil {
		return "", err
	}

	switch nodeType {
	case TypedefSimple:
		return md.getSimpleExpr(bPath)
	case TypedefCompact:
		return md.getCompactExpr(dd, bPath)
	case TypedefComplex:
		return md.getComplexExpr(dd, bPath)
	default:
		return "", fmt.Errorf("[getTypeExpr] unknown node type %##v %s", md[bPath], bPath)
	}
}

func (md Metadata) getVarName(dd *dsData, bPath string) (string, error) {
	if name := md[bPath].Name; name != "" {
		return name, nil
	}

	parentNode := md.getParentNode(bPath)
	prefix := parentNode.Name

	if prefix == "" {
		prefix = fmt.Sprintf("%s%d", parentNode.Prim, dd.counter)
		dd.counter++
	}

	suffix, err := md.getSuffix(dd, bPath)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", prefix, suffix), nil
}

func (md Metadata) getSuffix(dd *dsData, bPath string) (string, error) {
	parentNode := md.getParentNode(bPath)

	switch parentNode.Prim {
	case consts.LIST, consts.SET:
		return "_item", nil
	case consts.PAIR:
		return fmt.Sprintf("_arg%d", dd.arg), nil
	case consts.OR:
		return fmt.Sprintf("_var%d", dd.arg), nil
	case consts.OPTION:
		return "", nil
	case consts.MAP, consts.BIGMAP:
		if strings.HasSuffix(bPath, "k") {
			return "_key", nil
		}
		if strings.HasSuffix(bPath, "v") {
			return "_value", nil
		}
	}

	return "", fmt.Errorf("[getSuffix] error. prim: %s, bPath: %s", md[bPath].Prim, bPath)
}

func (md Metadata) getParentNode(bPath string) *NodeMetadata {
	parentPath := bPath[:len(bPath)-2]
	return md[parentPath]
}

func (md Metadata) getVarNameContractLambda(dd *dsData, bPath string) (string, error) {
	var node = md[bPath]
	var suffix string

	switch node.Type {
	case consts.CONTRACT:
		suffix = "_param"
	case consts.LAMBDA:
		suffix = "_input"
	default:
		return "", fmt.Errorf("[getVarNameContractLambda] error. node type: %s, bPath: %s", node.Type, bPath)
	}

	varName := fmt.Sprintf("%s%d%s", node.Type, dd.counter, suffix)
	dd.counter++

	return varName, nil
}

func (dd *dsData) insertTypedef(pos int, data Typedef) {
	dd.typedef = append(dd.typedef, Typedef{})
	copy(dd.typedef[pos+1:], dd.typedef[pos:])
	dd.typedef[pos] = data
}
