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

			for key := range m.All() {
				require.NotNil(t, key)
			}
		})
	}
}

func TestOrderedMap_Remove(t *testing.T) {
	str := func(value string) Node {
		return &String{Default: Default{Value: value, Prim: "string"}}
	}

	tests := []struct {
		name      string
		keys      []string
		remove    string
		wantOk    bool
		wantValue string
		wantKeys  []string
	}{
		{
			name:      "first",
			keys:      []string{"a", "b", "c"},
			remove:    "a",
			wantOk:    true,
			wantValue: "value_a",
			wantKeys:  []string{"b", "c"},
		}, {
			name:      "middle",
			keys:      []string{"a", "b", "c"},
			remove:    "b",
			wantOk:    true,
			wantValue: "value_b",
			wantKeys:  []string{"a", "c"},
		}, {
			name:      "last",
			keys:      []string{"a", "b", "c"},
			remove:    "c",
			wantOk:    true,
			wantValue: "value_c",
			wantKeys:  []string{"a", "b"},
		}, {
			name:      "single",
			keys:      []string{"a"},
			remove:    "a",
			wantOk:    true,
			wantValue: "value_a",
			wantKeys:  []string{},
		}, {
			name:     "absent",
			keys:     []string{"a", "b"},
			remove:   "z",
			wantOk:   false,
			wantKeys: []string{"a", "b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewOrderedMap()
			for _, key := range tt.keys {
				require.NoError(t, m.Add(str(key), str("value_"+key)))
			}

			removed, ok := m.Remove(str(tt.remove))
			require.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				res, err := removed.Compare(str(tt.wantValue))
				require.NoError(t, err)
				require.Equal(t, 0, res)
			}

			require.Equal(t, len(tt.wantKeys), m.Len())
			require.Len(t, m.values, len(tt.wantKeys))

			for i, key := range tt.wantKeys {
				res, err := m.keys[i].Compare(str(key))
				require.NoError(t, err)
				require.Equal(t, 0, res)

				value, ok := m.Get(str(key))
				require.True(t, ok, "key %s must remain in map", key)
				res, err = value.Compare(str("value_" + key))
				require.NoError(t, err)
				require.Equal(t, 0, res)
			}

			_, ok = m.Get(str(tt.remove))
			require.False(t, ok)
		})
	}
}
