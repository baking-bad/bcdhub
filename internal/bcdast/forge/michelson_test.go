package forge

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getStringPtr(val string) *string {
	return &val
}

func TestMichelson_Unforge(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    []*Node
		wantErr bool
	}{
		{
			name: "Small int",
			data: "0006",
			want: []*Node{
				{
					IntValue: big.NewInt(6),
				},
			},
		},
		{
			name: "Negative small int",
			data: "0046",
			want: []*Node{
				{
					IntValue: big.NewInt(-6),
				},
			},
		},
		{
			name: "Medium int",
			data: "00840e",
			want: []*Node{
				{
					IntValue: big.NewInt(900),
				},
			},
		},
		{
			name: "Negative medium int",
			data: "00c40e",
			want: []*Node{
				{
					IntValue: big.NewInt(-900),
				},
			},
		},
		{
			name: "Large int",
			data: "00ba9af7ea06",
			want: []*Node{
				{
					IntValue: big.NewInt(917431994),
				},
			},
		},
		{
			name: "Negative large int",
			data: "00c0f9b9d4c723",
			want: []*Node{
				{
					IntValue: big.NewInt(-610913435200),
				},
			},
		},
		{
			name: "String",
			data: "01000000096d696368656c696e65",
			want: []*Node{
				{
					StringValue: getStringPtr("micheline"),
				},
			},
		},
		{
			name: "Empty string",
			data: "0100000000",
			want: []*Node{
				{
					StringValue: getStringPtr(""),
				},
			},
		},
		{
			name: "Bytes",
			data: "0a000000080123456789abcdef",
			want: []*Node{
				{
					BytesValue: getStringPtr("0123456789abcdef"),
				},
			},
		},
		{
			name: "Mixed literal array",
			data: "02000000210061010000000574657a6f730100000000010000000b63727970746f6e6f6d6963",
			want: []*Node{
				{
					Prim: PrimArray,
					Args: []*Node{
						{
							IntValue: big.NewInt(-33),
						}, {
							StringValue: getStringPtr("tezos"),
						}, {
							StringValue: getStringPtr(""),
						}, {
							StringValue: getStringPtr("cryptonomic"),
						},
					},
				},
			},
		},
		{
			name: "Single primitive",
			data: "0343",
			want: []*Node{
				{
					Prim: "PUSH",
				},
			},
		},
		{
			name: "Single primitive with a single annotation",
			data: "04430000000440636261",
			want: []*Node{
				{
					Prim:   "PUSH",
					Annots: []string{"@cba"},
				},
			},
		},
		{
			name: "Single primitive with a single argument",
			data: "053d036d",
			want: []*Node{
				{
					Prim: "NIL",
					Args: []*Node{{
						Prim: "operation",
					}},
				},
			},
		},
		{
			name: "Single primitive with a single argument and annotation",
			data: "063d036d0000000440636261",
			want: []*Node{
				{
					Prim: "NIL",
					Args: []*Node{{
						Prim: "operation",
					}},
					Annots: []string{"@cba"},
				},
			},
		},
		{
			name: "Single primitive with two arguments",
			data: "073d036d036d",
			want: []*Node{
				{
					Prim: "NIL",
					Args: []*Node{{
						Prim: "operation",
					}, {
						Prim: "operation",
					}},
				},
			},
		},
		{
			name: "Single primitive with two arguments and annotation",
			data: "083d036d036d0000000440636261",
			want: []*Node{
				{
					Prim: "NIL",
					Args: []*Node{{
						Prim: "operation",
					}, {
						Prim: "operation",
					}},
					Annots: []string{"@cba"},
				},
			},
		},
		{
			name: "Single primitive with more than two arguments and no annotations",
			data: "093d00000006036d036d036d00000000",
			want: []*Node{
				{
					Prim: "NIL",
					Args: []*Node{{
						Prim: "operation",
					}, {
						Prim: "operation",
					}, {
						Prim: "operation",
					}},
					Annots: []string{},
				},
			},
		},
		{
			name: "Single primitive with more than two arguments and multiple annotations",
			data: "093d00000006036d036d036d00000011407265642040677265656e2040626c7565",
			want: []*Node{
				{
					Prim: "NIL",
					Args: []*Node{{
						Prim: "operation",
					}, {
						Prim: "operation",
					}, {
						Prim: "operation",
					}},
					Annots: []string{"@red", "@green", "@blue"},
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			data: "010000000561646d696e",
			want: []*Node{
				{
					StringValue: getStringPtr("admin"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			data: "0a00000016000006a868bd80219eb1f6a25108d1bdaa98ae27b2d9",
			want: []*Node{
				{
					BytesValue: getStringPtr("000006a868bd80219eb1f6a25108d1bdaa98ae27b2d9"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			data: "0100000006706175736564",
			want: []*Node{
				{
					StringValue: getStringPtr("paused"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			data: "0303",
			want: []*Node{
				{
					Prim: "False",
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			data: "010000000866616c6c6261636b",
			want: []*Node{
				{
					StringValue: getStringPtr("fallback"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			data: "02000000270316031607430368010000001655706172616d4e6f53756368456e747279506f696e7403420327",
			want: []*Node{
				{
					Prim: PrimArray,
					Args: []*Node{
						{
							Prim: "CAR",
						},
						{
							Prim: "CAR",
						},
						{
							Prim: "PUSH",
							Args: []*Node{
								{
									Prim: "string",
								}, {
									StringValue: getStringPtr("UparamNoSuchEntryPoint"),
								},
							},
						},
						{
							Prim: "PAIR",
						},
						{
							Prim: "FAILWITH",
						},
					},
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			data: "01000000086e65774f776e6572",
			want: []*Node{
				{
					StringValue: getStringPtr("newOwner"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			data: "0306",
			want: []*Node{
				{
					Prim: "None",
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			data: "01000000096f70657261746f7273",
			want: []*Node{
				{
					StringValue: getStringPtr("operators"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			data: "0200000000",
			want: []*Node{
				{
					Prim: PrimArray,
					Args: []*Node{},
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			data: "0100000009746f6b656e636f6465",
			want: []*Node{
				{
					StringValue: getStringPtr("tokencode"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			data: "0100000005545a425443",
			want: []*Node{
				{
					StringValue: getStringPtr("TZBTC"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			data: "0100000009746f6b656e6e616d65",
			want: []*Node{
				{
					StringValue: getStringPtr("tokenname"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			data: "010000000b746f74616c4275726e6564",
			want: []*Node{
				{
					StringValue: getStringPtr("totalBurned"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			data: "0000",
			want: []*Node{
				{
					IntValue: big.NewInt(0),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			data: "010000000b746f74616c4d696e746564",
			want: []*Node{
				{
					StringValue: getStringPtr("totalMinted"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			data: "010000000b746f74616c537570706c79",
			want: []*Node{
				{
					StringValue: getStringPtr("totalSupply"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			data: "010000000d72656465656d41646472657373",
			want: []*Node{
				{
					StringValue: getStringPtr("redeemAddress"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			data: "0a00000016000006a868bd80219eb1f6a25108d1bdaa98ae27b2d9",
			want: []*Node{
				{
					BytesValue: getStringPtr("000006a868bd80219eb1f6a25108d1bdaa98ae27b2d9"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMichelson()
			b, err := hex.DecodeString(tt.data)
			if err != nil {
				t.Errorf("Michelson.Unforge() DecodeString error = %v", err)
				return
			}
			_, err = m.Unforge(b)
			if (err != nil) != tt.wantErr {
				t.Errorf("Michelson.Unforge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(m.Nodes) != len(tt.want) {
				t.Errorf("Michelson.Unforge() len = %d, len(want) %v", len(m.Nodes), len(tt.want))
				return
			}
			for i := range tt.want {
				if !assert.Equal(t, tt.want[i], m.Nodes[i]) {
					return
				}
			}
		})
	}
}

func TestMichelson_Forge(t *testing.T) {
	tests := []struct {
		name    string
		nodes   []*Node
		want    string
		wantErr bool
	}{
		{
			name: "Small int",
			want: "0006",
			nodes: []*Node{
				{
					IntValue: big.NewInt(6),
				},
			},
		},
		{
			name: "Negative small int",
			want: "0046",
			nodes: []*Node{
				{
					IntValue: big.NewInt(-6),
				},
			},
		},
		{
			name: "Medium int",
			want: "00840e",
			nodes: []*Node{
				{
					IntValue: big.NewInt(900),
				},
			},
		},
		{
			name: "Negative medium int",
			want: "00c40e",
			nodes: []*Node{
				{
					IntValue: big.NewInt(-900),
				},
			},
		},
		{
			name: "Large int",
			want: "00ba9af7ea06",
			nodes: []*Node{
				{
					IntValue: big.NewInt(917431994),
				},
			},
		},
		{
			name: "Negative large int",
			want: "00c0f9b9d4c723",
			nodes: []*Node{
				{
					IntValue: big.NewInt(-610913435200),
				},
			},
		},
		{
			name: "String",
			want: "01000000096d696368656c696e65",
			nodes: []*Node{
				{
					StringValue: getStringPtr("micheline"),
				},
			},
		},
		{
			name: "Empty string",
			want: "0100000000",
			nodes: []*Node{
				{
					StringValue: getStringPtr(""),
				},
			},
		},
		{
			name: "Bytes",
			want: "0a000000080123456789abcdef",
			nodes: []*Node{
				{
					BytesValue: getStringPtr("0123456789abcdef"),
				},
			},
		},
		{
			name: "Mixed literal array",
			want: "02000000210061010000000574657a6f730100000000010000000b63727970746f6e6f6d6963",
			nodes: []*Node{
				{
					Prim: PrimArray,
					Args: []*Node{
						{
							IntValue: big.NewInt(-33),
						}, {
							StringValue: getStringPtr("tezos"),
						}, {
							StringValue: getStringPtr(""),
						}, {
							StringValue: getStringPtr("cryptonomic"),
						},
					},
				},
			},
		},
		{
			name: "Single primitive",
			want: "0343",
			nodes: []*Node{
				{
					Prim: "PUSH",
				},
			},
		},
		{
			name: "Single primitive with a single annotation",
			want: "04430000000440636261",
			nodes: []*Node{
				{
					Prim:   "PUSH",
					Annots: []string{"@cba"},
				},
			},
		},
		{
			name: "Single primitive with a single argument",
			want: "053d036d",
			nodes: []*Node{
				{
					Prim: "NIL",
					Args: []*Node{{
						Prim: "operation",
					}},
				},
			},
		},
		{
			name: "Single primitive with a single argument and annotation",
			want: "063d036d0000000440636261",
			nodes: []*Node{
				{
					Prim: "NIL",
					Args: []*Node{{
						Prim: "operation",
					}},
					Annots: []string{"@cba"},
				},
			},
		},
		{
			name: "Single primitive with two arguments",
			want: "073d036d036d",
			nodes: []*Node{
				{
					Prim: "NIL",
					Args: []*Node{{
						Prim: "operation",
					}, {
						Prim: "operation",
					}},
				},
			},
		},
		{
			name: "Single primitive with two arguments and annotation",
			want: "083d036d036d0000000440636261",
			nodes: []*Node{
				{
					Prim: "NIL",
					Args: []*Node{{
						Prim: "operation",
					}, {
						Prim: "operation",
					}},
					Annots: []string{"@cba"},
				},
			},
		},
		{
			name: "Single primitive with more than two arguments and multiple annotations",
			want: "093d00000006036d036d036d00000011407265642040677265656e2040626c7565",
			nodes: []*Node{
				{
					Prim: "NIL",
					Args: []*Node{{
						Prim: "operation",
					}, {
						Prim: "operation",
					}, {
						Prim: "operation",
					}},
					Annots: []string{"@red", "@green", "@blue"},
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			want: "010000000561646d696e",
			nodes: []*Node{
				{
					StringValue: getStringPtr("admin"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			want: "0a00000016000006a868bd80219eb1f6a25108d1bdaa98ae27b2d9",
			nodes: []*Node{
				{
					BytesValue: getStringPtr("000006a868bd80219eb1f6a25108d1bdaa98ae27b2d9"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			want: "0100000006706175736564",
			nodes: []*Node{
				{
					StringValue: getStringPtr("paused"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			want: "0303",
			nodes: []*Node{
				{
					Prim: "False",
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			want: "010000000866616c6c6261636b",
			nodes: []*Node{
				{
					StringValue: getStringPtr("fallback"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			want: "02000000270316031607430368010000001655706172616d4e6f53756368456e747279506f696e7403420327",
			nodes: []*Node{
				{
					Prim: PrimArray,
					Args: []*Node{
						{
							Prim: "CAR",
						},
						{
							Prim: "CAR",
						},
						{
							Prim: "PUSH",
							Args: []*Node{
								{
									Prim: "string",
								}, {
									StringValue: getStringPtr("UparamNoSuchEntryPoint"),
								},
							},
						},
						{
							Prim: "PAIR",
						},
						{
							Prim: "FAILWITH",
						},
					},
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			want: "01000000086e65774f776e6572",
			nodes: []*Node{
				{
					StringValue: getStringPtr("newOwner"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			want: "0306",
			nodes: []*Node{
				{
					Prim: "None",
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			want: "01000000096f70657261746f7273",
			nodes: []*Node{
				{
					StringValue: getStringPtr("operators"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			want: "0200000000",
			nodes: []*Node{
				{
					Prim: PrimArray,
					Args: []*Node{},
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			want: "0100000009746f6b656e636f6465",
			nodes: []*Node{
				{
					StringValue: getStringPtr("tokencode"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			want: "0100000005545a425443",
			nodes: []*Node{
				{
					StringValue: getStringPtr("TZBTC"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			want: "0100000009746f6b656e6e616d65",
			nodes: []*Node{
				{
					StringValue: getStringPtr("tokenname"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			want: "010000000b746f74616c4275726e6564",
			nodes: []*Node{
				{
					StringValue: getStringPtr("totalBurned"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			want: "0000",
			nodes: []*Node{
				{
					IntValue: big.NewInt(0),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			want: "010000000b746f74616c4d696e746564",
			nodes: []*Node{
				{
					StringValue: getStringPtr("totalMinted"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			want: "010000000b746f74616c537570706c79",
			nodes: []*Node{
				{
					StringValue: getStringPtr("totalSupply"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			want: "010000000d72656465656d41646472657373",
			nodes: []*Node{
				{
					StringValue: getStringPtr("redeemAddress"),
				},
			},
		},
		{
			name: "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			want: "0a00000016000006a868bd80219eb1f6a25108d1bdaa98ae27b2d9",
			nodes: []*Node{
				{
					BytesValue: getStringPtr("000006a868bd80219eb1f6a25108d1bdaa98ae27b2d9"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Michelson{
				Nodes: tt.nodes,
			}
			got, err := m.Forge()
			if (err != nil) != tt.wantErr {
				t.Errorf("Michelson.Forge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotStr := hex.EncodeToString(got)
			if gotStr != tt.want {
				t.Errorf("Michelson.Forge() = %v, want %v", gotStr, tt.want)
			}
		})
	}
}
