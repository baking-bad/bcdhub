package ast

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
)

// docs constants
const (
	DocsFull = ""
)

var simpleTypes = []string{
	consts.INT,
	consts.STRING,
	consts.BYTES,
	consts.BOOL,
	consts.NAT,
	consts.MUTEZ,
	consts.TIMESTAMP,
	consts.ADDRESS,
	consts.KEYHASH,
	consts.KEY,
	consts.SIGNATURE,
	consts.CHAINID,
	consts.UNIT,
	consts.OPERATION,
	consts.BAKERHASH,
	consts.BLS12381FR,
	consts.BLS12381G1,
	consts.BLS12381G2,
	consts.NEVER,
	consts.SAPLINGSTATE,
	consts.SAPLINGTRANSACTION,
}

// EntrypointType -
type EntrypointType struct {
	Name string    `json:"name"`
	Type []Typedef `json:"typedef"`
}

// Typedef -
type Typedef struct {
	Name    string       `json:"name"`
	Type    string       `json:"type,omitempty"`
	TypeDef []Typedef    `json:"typedef,omitempty"`
	Args    []TypedefArg `json:"args,omitempty"`
}

// TypedefArg -
type TypedefArg struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value"`
}

func buildArrayDocs(nodes []Node) ([]Typedef, error) {
	typedef := make([]Typedef, 0)
	for i := range nodes {
		docs, _, err := nodes[i].Docs("")
		if err != nil {
			return nil, err
		}
		typedef = append(typedef, docs...)
	}
	return typedef, nil
}

func isSimpleDocType(prim string) bool {
	for _, t := range simpleTypes {
		if prim == t {
			return true
		}
	}

	return false
}

func makeVarDocString(name string) string {
	return fmt.Sprintf("$%s", strings.TrimPrefix(name, "@"))
}

func getNameDocString(node Type, inferredName string) string {
	name := node.GetName()
	if strings.HasPrefix(name, "@") && inferredName != "" {
		name = inferredName
	}
	return name
}

func isFlatDocType(typ Typedef) bool {
	if isSyntheticDocName(typ) {
		return false
	}
	return strings.HasPrefix(typ.Type, consts.LIST) ||
		strings.HasPrefix(typ.Type, consts.SET) ||
		strings.HasPrefix(typ.Type, consts.BIGMAP) ||
		strings.HasPrefix(typ.Type, consts.CONTRACT) ||
		strings.HasPrefix(typ.Type, consts.MAP) ||
		strings.HasPrefix(typ.Type, consts.OPTION)
}

func isSyntheticDocName(typ Typedef) bool {
	return strings.HasSuffix(typ.Name, "_param") ||
		strings.HasSuffix(typ.Name, "_item") ||
		strings.HasSuffix(typ.Name, "_key") ||
		strings.HasSuffix(typ.Name, "_value") ||
		typ.Name == "input" ||
		typ.Name == "return"
}
