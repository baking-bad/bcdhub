package contractparser

import (
	"log"
	"strings"

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

func newNodeJSON(data gjson.Result) Node {
	if !data.IsObject() {
		log.Panicf("Unknown node type: %v", data)
	}
	n := Node{
		Child:       make([]Node, 0),
		Annotations: make([]string, 0),
		Prim:        strings.ToUpper(data.Get(keyPrim).String()),
		Args:        data.Get(keyArgs),
	}
	for _, a := range data.Get(keyAnnots).Array() {
		n.Annotations = append(n.Annotations, a.String())
	}

	if data.Get(keyBytes).Exists() {
		n.Value = data.Get(keyBytes).String()
		n.Type = keyBytes
	} else if data.Get(keyInt).Exists() {
		n.Value = data.Get(keyInt).Int()
		n.Type = keyInt
	} else if data.Get(keyMutez).Exists() {
		n.Value = data.Get(keyMutez).Int()
		n.Type = keyMutez
	} else if data.Get(keyString).Exists() {
		n.Value = data.Get(keyString).String()
		n.Type = keyString
	} else if data.Get(keyTime).Exists() {
		n.Value = data.Get(keyTime).String()
		n.Type = keyTime
	}

	return n
}

// GetString - return string value
func (n *Node) GetString() string {
	if n.Type != "string" {
		return ""
	}
	return n.Value.(string)
}

// GetBytes - return bytes value
func (n *Node) GetBytes() string {
	if n.Type != "bytes" {
		return ""
	}
	return n.Value.(string)
}

// GetInt - return int value
func (n *Node) GetInt() int64 {
	if n.Type != "int" {
		return 0
	}
	return n.Value.(int64)
}

// GetMutez - return mutez value
func (n *Node) GetMutez() float64 {
	if n.Type != "mutez" {
		return .0
	}
	return n.Value.(float64)
}

// Is - check prim value
func (n *Node) Is(prim string) bool {
	return strings.ToLower(prim) == n.Prim
}

// HasAnnots - check if node has annotations
func (n *Node) HasAnnots() bool {
	return len(n.Annotations) > 0
}

// HasArgs - check if node has args
func (n *Node) HasArgs() bool {
	return len(n.Args.Array()) > 0
}

// Print -
func (n *Node) Print() {
	log.Printf("%s: %s [%s] (args: %d)", strings.Join(n.Annotations, ","), n.Prim, n.Type, len(n.Args.Array()))
}
