package node

import (
	"log"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/tidwall/gjson"
)

// Node -
type Node struct {
	Prim        string       `json:"prim,omitempty"`
	Args        gjson.Result `json:"args,omitempty"`
	Annotations []string     `json:"annots,omitempty"`
	Value       interface{}  `json:"value,omitempty"`
	Type        string       `json:"type,omitempty"`
	Child       []Node       `json:"child,omitempty"`
	Path        string       `json:"-"`
}

// NewNodeJSON -
func NewNodeJSON(data gjson.Result) Node {
	if !data.IsObject() {
		log.Panicf("Unknown node type: %v", data)
	}
	n := Node{
		Child:       make([]Node, 0),
		Annotations: make([]string, 0),
		Prim:        strings.ToLower(data.Get(consts.KeyPrim).String()),
		Args:        data.Get(consts.KeyArgs),
	}
	for _, a := range data.Get(consts.KeyAnnots).Array() {
		n.Annotations = append(n.Annotations, a.String())
	}

	if data.Get(consts.KeyBytes).Exists() {
		n.Value = data.Get(consts.KeyBytes).String()
		n.Type = consts.KeyBytes
	} else if data.Get(consts.KeyInt).Exists() {
		n.Value = data.Get(consts.KeyInt).Int()
		n.Type = consts.KeyInt
	} else if data.Get(consts.KeyMutez).Exists() {
		n.Value = data.Get(consts.KeyMutez).Int()
		n.Type = consts.KeyMutez
	} else if data.Get(consts.KeyString).Exists() {
		n.Value = data.Get(consts.KeyString).String()
		n.Type = consts.KeyString
	} else if data.Get(consts.KeyTime).Exists() {
		n.Value = data.Get(consts.KeyTime).Time()
		n.Type = consts.KeyTime
	}

	return n
}

// GetString - return string value
func (n Node) GetString() string {
	if n.Type != consts.KeyString {
		return ""
	}
	return n.Value.(string)
}

// GetBytes - return bytes value
func (n Node) GetBytes() string {
	if n.Type != consts.KeyBytes {
		return ""
	}
	return n.Value.(string)
}

// GetInt - return int value
func (n Node) GetInt() int64 {
	if n.Type != consts.KeyInt {
		return 0
	}
	return n.Value.(int64)
}

// GetMutez - return mutez value
func (n Node) GetMutez() float64 {
	if n.Type != consts.KeyMutez {
		return .0
	}
	return n.Value.(float64)
}

// Is - check prim value
func (n Node) Is(prim string) bool {
	return strings.ToLower(prim) == n.Prim
}

// HasAnnots - check if node has annotations
func (n Node) HasAnnots() bool {
	return len(n.Annotations) > 0
}

// HasArgs - check if node has args
func (n Node) HasArgs() bool {
	return len(n.Args.Array()) > 0
}

// Print -
func (n Node) Print() {
	log.Printf("%s: %s [%s] (args: %d)", strings.Join(n.Annotations, ","), n.Prim, n.Type, len(n.Args.Array()))
}
