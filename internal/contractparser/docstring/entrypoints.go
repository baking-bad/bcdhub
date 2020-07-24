package docstring

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
)

// EntrypointType -
type EntrypointType struct {
	Name    string    `json:"name"`
	Type    []Typedef `json:"typedef"`
	BinPath string    `json:"bin_path"`
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

// GetTypedef -
func GetTypedef(binPath string, md meta.Metadata) ([]Typedef, error) {
	if root := md[binPath]; root.Name == "" {
		root.Name = root.GetEntrypointName(-1)
	}

	return parseEntrypointTypes(binPath, md)
}

// GetEntrypoints -
func GetEntrypoints(md meta.Metadata) ([]EntrypointType, error) {
	root := md["0"]
	entrypoints := make([]EntrypointType, 0)

	if md.IsComplexEntryRoot() {
		for i, binPath := range root.Args {
			md[binPath].Name = md[binPath].GetEntrypointName(i)

			typeDefs, err := parseEntrypointTypes(binPath, md)
			if err != nil {
				return nil, err
			}

			entrypoints = append(entrypoints, EntrypointType{
				Name:    md[binPath].Name,
				Type:    typeDefs,
				BinPath: binPath,
			})
		}
	} else {
		root.Name = root.GetEntrypointName(-1)
		typeDefs, err := parseEntrypointTypes("0", md)
		if err != nil {
			return nil, err
		}

		entrypoints = append(entrypoints, EntrypointType{
			Name:    root.Name,
			Type:    typeDefs,
			BinPath: "0",
		})
	}

	return entrypoints, nil
}

func parseEntrypointTypes(bPath string, md meta.Metadata) ([]Typedef, error) {
	var dd dsData

	typ, err := getTypeExpr(&dd, bPath, md)
	if err != nil {
		return nil, err
	}

	node := md[bPath]

	if isSimpleType(node.Prim) || isCompactType(bPath, md) {
		dd.insertTypedef(0, Typedef{
			Name: node.Name,
			Type: typ,
		})
	}

	return dd.typedef, nil
}

func handleType(dd *dsData, bPath string, md meta.Metadata) (string, error) {
	i := len(dd.typedef)

	switch md[bPath].Type {
	case consts.TypeTuple, consts.TypeEnum, consts.TypeUnion:
		return handleTupleEnumUnion(dd, bPath, i, md)
	case consts.TypeNamedTuple, consts.TypeNamedEnum, consts.TypeNamedUnion:
		return handleNamed(dd, bPath, i, md)
	case consts.CONTRACT:
		return handleContract(dd, bPath, i, md)
	case consts.MAP, consts.BIGMAP:
		return handleMap(dd, bPath, i, md)
	case consts.LAMBDA:
		return handleLambda(dd, bPath, i, md)
	default:
		return "", fmt.Errorf("[handleType] %##v %s", md[bPath], bPath)
	}
}

func getTypeExpr(dd *dsData, bPath string, md meta.Metadata) (string, error) {
	nodeType, err := getType(bPath, md)
	if err != nil {
		return "", err
	}

	switch nodeType {
	case TypedefSimple:
		return getSimpleExpr(bPath, md)
	case TypedefCompact:
		return getCompactExpr(dd, bPath, md)
	case TypedefComplex:
		return getComplexExpr(dd, bPath, md)
	default:
		return "", fmt.Errorf("[getTypeExpr] unknown node type %##v %s", md[bPath], bPath)
	}
}

func getVarName(dd *dsData, bPath string, md meta.Metadata) (string, error) {
	if name := md[bPath].Name; name != "" {
		return name, nil
	}

	parentNode := getParentNode(bPath, md)
	prefix := parentNode.Name

	if prefix == "" {
		prefix = fmt.Sprintf("%s%d", parentNode.Prim, dd.counter)
		dd.counter++
	}

	suffix, err := getSuffix(dd, bPath, md)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", prefix, suffix), nil
}

func getSuffix(dd *dsData, bPath string, md meta.Metadata) (string, error) {
	parentNode := getParentNode(bPath, md)

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

func getParentNode(bPath string, md meta.Metadata) *meta.NodeMetadata {
	parentPath := bPath[:len(bPath)-2]
	return md[parentPath]
}

func getVarNameContractLambda(dd *dsData, bPath string, md meta.Metadata) (string, error) {
	var node = md[bPath]
	var suffix string

	switch node.Type {
	case consts.CONTRACT:
		suffix = "_param"
	case consts.LAMBDA:
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
