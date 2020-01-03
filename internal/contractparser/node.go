package contractparser

import (
	"log"
	"strings"
)

// Node -
type Node struct {
	Prim        string
	Args        []interface{}
	Annotations []string
	Value       interface{}
	Type        string
}

func newNode(obj map[string]interface{}) *Node {
	n := Node{}
	for k, v := range obj {
		switch k {
		case "args":
			if vargs, ok := v.([]interface{}); ok {
				n.Args = vargs
			}
		case "prim":
			n.Prim = strings.ToUpper(v.(string))
		case "annots":
			n.Annotations = make([]string, 0)
			annots := v.([]interface{})
			for i := range annots {
				n.Annotations = append(n.Annotations, strings.Trim(annots[i].(string), "%@"))
			}
		case "string", "int", "mutez", "bytes":
			n.Value = v
			n.Type = k
		default:
			log.Printf("Unknown node key: %s", k)
		}

	}
	return &n
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
	return strings.ToUpper(prim) == n.Prim
}

// HasAnnots - check if node has annotations
func (n *Node) HasAnnots() bool {
	return len(n.Annotations) > 0
}
