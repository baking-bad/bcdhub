package contractparser

import (
	"testing"

	"github.com/tidwall/gjson"
)

func TestIsDelegatorContract(t *testing.T) {
	tests := []struct {
		name string
		data string
		want bool
	}{
		{
			name: "Case 1: KT1K5KsAoXkShk2m31Cw3fzmKRPkDKecHHEJ",
			data: `{"code":[{"prim":"parameter","args":[{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%do"]},{"prim":"unit","annots":["%default"]}]}]},{"prim":"storage","args":[{"prim":"key_hash"}]},{"prim":"code","args":[[[[{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]}]],{"prim":"IF_LEFT","args":[[{"prim":"PUSH","args":[{"prim":"mutez"},{"int":"0"}]},{"prim":"AMOUNT"},[[{"prim":"COMPARE"},{"prim":"EQ"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],[{"prim":"DIP","args":[[{"prim":"DUP"}]]},{"prim":"SWAP"}],{"prim":"IMPLICIT_ACCOUNT"},{"prim":"ADDRESS"},{"prim":"SENDER"},[[{"prim":"COMPARE"},{"prim":"EQ"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"EXEC"},{"prim":"PAIR"}],[{"prim":"DROP"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]]}],"storage":{"string":"tz1f5FMfgXgMhkBuw4f2Drq5KUbsLZLEN8J5"}}`,
			want: true,
		}, {
			name: "Case 2: KT1T14WrhNCvaBDwn1TxxnseFDvFJCUJuJkD",
			data: ``,
			want: true,
		}, {
			name: "Case 3: KT1T14WrhNCvaBDwn1TxxnseFDvFJCUJuJkD",
			data: `[]`,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := gjson.Parse(tt.data)
			if got := IsDelegatorContract(data); got != tt.want {
				t.Errorf("IsDelegatorContract() = %v, want %v", got, tt.want)
			}
		})
	}
}
