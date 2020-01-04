package contractparser

import (
	"encoding/json"
	"log"
)

// Schema -
type Schema []Node

func newSchema(v interface{}) Schema {
	s := make(Schema, 0)
	switch val := v.(type) {
	case []interface{}:
		for _, a := range val {
			s = append(s, newSchema(a)...)
		}
	case map[string]interface{}:
		n := newNode(val)
		if n.Is("pair") || n.Is("or") {
			for i := range n.Child {
				s = append(s, n.Child[i])
			}
		} else {
			s = append(s, n)
		}
	}
	return s
}

// Print -
func (s Schema) Print() {
	if b, err := json.MarshalIndent(s, "", " "); err == nil {
		log.Print(string(b))
	} else {
		log.Println(err)
	}
}
