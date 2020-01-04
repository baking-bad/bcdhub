package contractparser

import (
	"log"
	"strings"
)

// Node -
type Node struct {
	Prim        string        `json:"prim,omitempty"`
	Args        []interface{} `json:"-"`
	Annotations []string      `json:"annots,omitempty"`
	Value       interface{}   `json:"value,omitempty"`
	Type        string        `json:"type,omitempty"`
	Child       []Node        `json:"child,omitempty"`
}

func newNode(obj map[string]interface{}) Node {
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
				n.Annotations = append(n.Annotations, strings.ToLower(strings.Trim(annots[i].(string), "%@:")))
			}
		case "string", "int", "mutez", "bytes":
			n.Value = v
			n.Type = k
		default:
			log.Printf("Unknown node key: %s", k)
		}

	}

	n.getChild()
	return n
}

func (n *Node) getChild() {
	for _, arg := range n.Args {
		switch a := arg.(type) {
		case []interface{}:
			empty := newNode(map[string]interface{}{
				"args": a,
			})
			n.Child = append(n.Child, empty)
		case map[string]interface{}:
			c := newNode(a)
			switch c.Prim {
			case "PAIR", "OR":
				for i := range c.Child {
					n.Child = append(n.Child, c.Child[i])
				}
			default:
				n.Child = append(n.Child, c)
			}
		default:
			log.Printf("getChild: Unknown type %T", a)
		}
	}
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

// HasArgs - check if node has args
func (n *Node) HasArgs() bool {
	return len(n.Args) > 0
}

// Print -
func (n *Node) Print() {
	log.Printf("%s: %s [%s] (args: %d)", strings.Join(n.Annotations, ","), n.Prim, n.Type, len(n.Args))
}
