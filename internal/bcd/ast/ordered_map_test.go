package ast

import (
	"testing"
)

func TestOrderedMap_Add(t *testing.T) {
	type item struct {
		key   Node
		value Node
	}
	tests := []struct {
		name              string
		data              []item
		wantKeys          []Node
		getKey            Node
		wantForGet        Node
		remove            Node
		lengthAfterRemove int
		wantErr           bool
	}{
		{
			name: "first",
			data: []item{
				{
					key:   &String{Default: Default{Value: "10", Prim: "string"}},
					value: &String{Default: Default{Value: "10", Prim: "string"}},
				}, {
					key:   &String{Default: Default{Value: "9", Prim: "string"}},
					value: &String{Default: Default{Value: "9", Prim: "string"}},
				}, {
					key:   &String{Default: Default{Value: "2", Prim: "string"}},
					value: &String{Default: Default{Value: "2", Prim: "string"}},
				}, {
					key:   &String{Default: Default{Value: "4", Prim: "string"}},
					value: &String{Default: Default{Value: "4", Prim: "string"}},
				}, {
					key:   &String{Default: Default{Value: "4", Prim: "string"}},
					value: &String{Default: Default{Value: "100", Prim: "string"}},
				},
			},
			wantKeys: []Node{
				&String{Default: Default{Value: "10", Prim: "string"}},
				&String{Default: Default{Value: "2", Prim: "string"}},
				&String{Default: Default{Value: "4", Prim: "string"}},
				&String{Default: Default{Value: "9", Prim: "string"}},
			},
			getKey:            &String{Default: Default{Value: "4", Prim: "string"}},
			wantForGet:        &String{Default: Default{Value: "100", Prim: "string"}},
			remove:            &String{Default: Default{Value: "2", Prim: "string"}},
			lengthAfterRemove: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewOrderedMap()
			for i := range tt.data {
				if err := m.Add(tt.data[i].key, tt.data[i].value); (err != nil) != tt.wantErr {
					t.Errorf("OrderedMap.Add() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}

			if len(tt.wantKeys) != m.Len() {
				t.Errorf("OrderedMap.Len() len(tt.wantKeys) = %d, m.Len() = %d", len(tt.wantKeys), m.Len())
				return
			}

			for i := range tt.wantKeys {
				res, err := tt.wantKeys[i].Compare(m.keys[i])
				if err != nil {
					t.Errorf("Compare err = %v", err)
					return
				}
				if res != 0 {
					t.Errorf("Compare res=%d, i=%d", res, i)
					return
				}
			}

			receive, ok := m.Get(tt.getKey)
			if !ok {
				t.Errorf("Get ok = %v", ok)
				return
			}
			res, err := receive.Compare(tt.wantForGet)
			if err != nil {
				t.Errorf("receive.Compare err = %v", err)
				return
			}
			if res != 0 {
				t.Errorf("receive.Compare res=%d", res)
				return
			}

			if _, ok := m.Remove(tt.remove); !ok {
				t.Errorf("OrderedMap.Remove() ok=%v", ok)
				return
			}
			if tt.lengthAfterRemove != m.Len() {
				t.Errorf("lengthAfterRemove=%d, m.Len()=%d", tt.lengthAfterRemove, m.Len())
				return
			}

			_ = m.Range(func(key, value Comparable) (bool, error) {
				return false, nil
			})
		})
	}
}
