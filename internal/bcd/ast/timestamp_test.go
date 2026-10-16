package ast

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/stretchr/testify/require"
)

func newTimestampFromString(t *testing.T, tree string) *TypedAst {
	t.Helper()
	typ, err := NewTypedAstFromString(tree)
	require.NoError(t, err)
	return typ
}

func TestTimestamp_ParseValue(t *testing.T) {
	tests := []struct {
		name        string
		node        string
		wantSeconds string
		wantDisplay string
		wantErr     bool
	}{
		{
			name:        "zero int",
			node:        `{"int":"0"}`,
			wantSeconds: "0",
			wantDisplay: "1970-01-01T00:00:00Z",
		}, {
			name:        "int",
			node:        `{"int":"1624101720"}`,
			wantSeconds: "1624101720",
			wantDisplay: "2021-06-19T11:22:00Z",
		}, {
			name:        "negative int",
			node:        `{"int":"-1"}`,
			wantSeconds: "-1",
			wantDisplay: "1969-12-31T23:59:59Z",
		}, {
			name:        "int is not treated as milliseconds",
			node:        `{"int":"1624101720000"}`,
			wantSeconds: "1624101720000",
			wantDisplay: "1624101720000",
		}, {
			name:        "int greater than int64",
			node:        `{"int":"11111111111111111111111111"}`,
			wantSeconds: "11111111111111111111111111",
			wantDisplay: "11111111111111111111111111",
		}, {
			name:        "max RFC3339",
			node:        `{"int":"253402300799"}`,
			wantSeconds: "253402300799",
			wantDisplay: "9999-12-31T23:59:59Z",
		}, {
			name:        "after max RFC3339",
			node:        `{"int":"253402300800"}`,
			wantSeconds: "253402300800",
			wantDisplay: "253402300800",
		}, {
			name:        "min RFC3339",
			node:        `{"int":"-62135596800"}`,
			wantSeconds: "-62135596800",
			wantDisplay: "0001-01-01T00:00:00Z",
		}, {
			name:        "before min RFC3339",
			node:        `{"int":"-62135596801"}`,
			wantSeconds: "-62135596801",
			wantDisplay: "-62135596801",
		}, {
			name:        "RFC3339 string",
			node:        `{"string":"2021-06-19T11:22:00Z"}`,
			wantSeconds: "1624101720",
			wantDisplay: "2021-06-19T11:22:00Z",
		}, {
			name:        "RFC3339 string with offset",
			node:        `{"string":"2021-06-19T14:22:00+03:00"}`,
			wantSeconds: "1624101720",
			wantDisplay: "2021-06-19T11:22:00Z",
		}, {
			name:        "numeric string",
			node:        `{"string":"1624101720"}`,
			wantSeconds: "1624101720",
			wantDisplay: "2021-06-19T11:22:00Z",
		}, {
			name:    "invalid string",
			node:    `{"string":"not-a-date"}`,
			wantErr: true,
		}, {
			name:    "empty string",
			node:    `{"string":""}`,
			wantErr: true,
		}, {
			name:    "bytes",
			node:    `{"bytes":"00"}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node base.Node
			require.NoError(t, json.UnmarshalFromString(tt.node, &node))

			ts := NewTimestamp(0)
			err := ts.ParseValue(&node)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			value, ok := ts.Value.(*types.BigInt)
			require.True(t, ok, "value must be *types.BigInt, got %T", ts.Value)
			require.Equal(t, tt.wantSeconds, value.String())

			miguel, err := ts.ToMiguel()
			require.NoError(t, err)
			require.Equal(t, tt.wantDisplay, miguel.Value)

			model := make(JSONModel)
			ts.GetJSONModel(model)
			require.Equal(t, tt.wantDisplay, model[ts.GetName()])
		})
	}
}

func TestTimestamp_ToBaseNode(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		optimized bool
		want      string
	}{
		{
			name: "readable",
			data: `{"int":"1624101720"}`,
			want: `{"string":"2021-06-19T11:22:00Z"}`,
		}, {
			name:      "optimized",
			data:      `{"string":"2021-06-19T11:22:00Z"}`,
			optimized: true,
			want:      `{"int":"1624101720"}`,
		}, {
			name: "readable out of RFC3339 range falls back to int",
			data: `{"int":"11111111111111111111111111"}`,
			want: `{"int":"11111111111111111111111111"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := newTimestampFromString(t, `{"prim":"timestamp"}`)
			require.NoError(t, typ.SettleFromBytes([]byte(tt.data)))

			node, err := typ.ToBaseNode(tt.optimized)
			require.NoError(t, err)
			got, err := json.MarshalToString(node)
			require.NoError(t, err)
			require.JSONEq(t, tt.want, got)
		})
	}
}

func TestTimestamp_FromJSONSchema(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    string
		wantErr bool
	}{
		{
			name:  "RFC3339",
			value: "2021-06-19T11:22:00Z",
			want:  `{"int":"1624101720"}`,
		}, {
			name:  "numeric string",
			value: "1624101720",
			want:  `{"int":"1624101720"}`,
		}, {
			name:  "number",
			value: 1624101720.0,
			want:  `{"int":"1624101720"}`,
		}, {
			name:  "negative number",
			value: -1.0,
			want:  `{"int":"-1"}`,
		}, {
			name:  "null",
			value: nil,
			want:  `{"int":null}`,
		}, {
			name:    "fractional number",
			value:   1.5,
			wantErr: true,
		}, {
			name:    "invalid string",
			value:   "tomorrow",
			wantErr: true,
		}, {
			name:    "bool",
			value:   true,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := newTimestampFromString(t, `{"prim":"timestamp","annots":["%ts"]}`)
			err := typ.FromJSONSchema(map[string]interface{}{"ts": tt.value})
			if tt.wantErr {
				require.ErrorIs(t, err, consts.ErrValidation)
				return
			}
			require.NoError(t, err)

			params, err := typ.ToParameters("")
			require.NoError(t, err)
			require.Equal(t, tt.want, string(params))
		})
	}
}

func TestTimestamp_ToParametersUnset(t *testing.T) {
	typ := newTimestampFromString(t, `{"prim":"timestamp","annots":["%ts"]}`)
	_, err := typ.ToParameters("")
	require.ErrorIs(t, err, consts.ErrValidation)

	_, err = typ.ToBaseNode(false)
	require.ErrorIs(t, err, consts.ErrValidation)
}

func TestTimestamp_Compare(t *testing.T) {
	settled := func(data string) *Timestamp {
		typ := newTimestampFromString(t, `{"prim":"timestamp"}`)
		require.NoError(t, typ.SettleFromBytes([]byte(data)))
		return typ.Nodes[0].(*Timestamp)
	}
	fromJSON := func(value interface{}) *Timestamp {
		typ := newTimestampFromString(t, `{"prim":"timestamp","annots":["%ts"]}`)
		require.NoError(t, typ.FromJSONSchema(map[string]interface{}{"ts": value}))
		return typ.Nodes[0].(*Timestamp)
	}

	tests := []struct {
		name    string
		x, y    Comparable
		want    int
		wantErr bool
	}{
		{
			name: "int and RFC3339 string are equal",
			x:    settled(`{"int":"1624101720"}`),
			y:    settled(`{"string":"2021-06-19T11:22:00Z"}`),
			want: 0,
		}, {
			name: "settled and JSON values are comparable",
			x:    settled(`{"int":"1624101720"}`),
			y:    fromJSON("2021-06-19T11:22:01Z"),
			want: -1,
		}, {
			name: "negative is less than zero",
			x:    settled(`{"int":"0"}`),
			y:    settled(`{"int":"-1"}`),
			want: 1,
		}, {
			name: "beyond int64",
			x:    settled(`{"int":"11111111111111111111111111"}`),
			y:    settled(`{"int":"11111111111111111111111112"}`),
			want: -1,
		}, {
			name:    "unset value",
			x:       NewTimestamp(0),
			y:       settled(`{"int":"1"}`),
			wantErr: true,
		}, {
			name:    "null value",
			x:       fromJSON(nil),
			y:       settled(`{"int":"1"}`),
			wantErr: true,
		}, {
			name:    "different type",
			x:       settled(`{"int":"1"}`),
			y:       NewNat(0),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.x.Compare(tt.y)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestTimestamp_MapKeyFromJSONSchema(t *testing.T) {
	typ := newTimestampFromString(t, `{"prim":"map","args":[{"prim":"timestamp"},{"prim":"nat"}],"annots":["%m"]}`)
	err := typ.FromJSONSchema(map[string]interface{}{
		"m": []interface{}{
			map[string]interface{}{"@timestamp_2": "2021-01-01T00:00:00Z", "@nat_3": 2.0},
			map[string]interface{}{"@timestamp_2": "2020-01-01T00:00:00Z", "@nat_3": 1.0},
			map[string]interface{}{"@timestamp_2": "2021-01-01T00:00:00Z", "@nat_3": 3.0},
		},
	})
	require.NoError(t, err)

	params, err := typ.ToParameters("")
	require.NoError(t, err)
	require.Equal(t, `[{"prim":"Elt","args":[{"int":"1577836800"},{"int":"1"}]},{"prim":"Elt","args":[{"int":"1609459200"},{"int":"3"}]}]`, string(params))
}

func TestTimestamp_GetJSONModelUnsettled(t *testing.T) {
	typ := newTimestampFromString(t, `{"prim":"pair","args":[{"prim":"timestamp","annots":["%ts"]},{"prim":"nat","annots":["%n"]}]}`)
	model := make(JSONModel)
	require.NotPanics(t, func() {
		typ.GetJSONModel(model)
	})
	require.Equal(t, "", model["ts"])
}

func TestTimestamp_Distinguish(t *testing.T) {
	tests := []struct {
		name         string
		current      string
		prev         string
		wantValue    interface{}
		wantFrom     interface{}
		wantDiffType string
	}{
		{
			name:      "same value",
			current:   `{"int":"1624101720"}`,
			prev:      `{"string":"2021-06-19T11:22:00Z"}`,
			wantValue: "2021-06-19T11:22:00Z",
		}, {
			name:         "update",
			current:      `{"int":"1624101721"}`,
			prev:         `{"int":"1624101720"}`,
			wantValue:    "2021-06-19T11:22:01Z",
			wantFrom:     "2021-06-19T11:22:00Z",
			wantDiffType: MiguelKindUpdate,
		}, {
			name:         "update out of RFC3339 range",
			current:      `{"int":"253402300800"}`,
			prev:         `{"int":"0"}`,
			wantValue:    "253402300800",
			wantFrom:     "1970-01-01T00:00:00Z",
			wantDiffType: MiguelKindUpdate,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := newTimestampFromString(t, `{"prim":"timestamp"}`)
			require.NoError(t, current.SettleFromBytes([]byte(tt.current)))
			prev := newTimestampFromString(t, `{"prim":"timestamp"}`)
			require.NoError(t, prev.SettleFromBytes([]byte(tt.prev)))

			diff, err := current.Diff(prev)
			require.NoError(t, err)
			require.Equal(t, tt.wantValue, diff.Value)
			require.Equal(t, tt.wantFrom, diff.From)
			require.Equal(t, tt.wantDiffType, diff.DiffType)
		})
	}
}
