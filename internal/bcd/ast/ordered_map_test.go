package ast

import (
	"testing"

	"github.com/stretchr/testify/require"
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
				err := m.Add(tt.data[i].key, tt.data[i].value)
				require.NoError(t, err)
			}

			require.Equal(t, len(tt.wantKeys), m.Len())

			for i := range tt.wantKeys {
				res, err := tt.wantKeys[i].Compare(m.keys[i])
				require.NoError(t, err)
				require.Equal(t, res, 0)
			}

			receive, ok := m.Get(tt.getKey)
			require.True(t, ok)

			res, err := receive.Compare(tt.wantForGet)
			require.NoError(t, err)
			require.Equal(t, res, 0)

			_, ok = m.Remove(tt.remove)
			require.True(t, ok)

			require.Equal(t, tt.lengthAfterRemove, m.Len())

			_ = m.Range(func(key, value Comparable) (bool, error) {
				return false, nil
			})
		})
	}
}
