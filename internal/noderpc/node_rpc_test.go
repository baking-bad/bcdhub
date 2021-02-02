package noderpc

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/tidwall/gjson"
)

func getHandler(status int, response string) func(*http.Request) (*http.Response, error) {
	return func(*http.Request) (*http.Response, error) {
		r := ioutil.NopCloser(bytes.NewReader([]byte(response)))
		return &http.Response{
			StatusCode: status,
			Body:       r,
		}, nil
	}
}

func TestNodeRPC_GetScriptJSON(t *testing.T) {
	type fields struct {
		legacy     bool
		statusCode int
	}
	tests := []struct {
		name    string
		fields  fields
		script  string
		wantErr bool
	}{
		{
			name:   "legacy",
			script: `{"code":[{"prim":"parameter","args":[{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%do"]},{"prim":"unit","annots":["%default"]}]}]},{"prim":"storage","args":[{"prim":"key_hash"}]},{"prim":"code","args":[[[[{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]}]],{"prim":"IF_LEFT","args":[[{"prim":"PUSH","args":[{"prim":"mutez"},{"int":"0"}]},{"prim":"AMOUNT"},[[{"prim":"COMPARE"},{"prim":"EQ"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],[{"prim":"DIP","args":[[{"prim":"DUP"}]]},{"prim":"SWAP"}],{"prim":"IMPLICIT_ACCOUNT"},{"prim":"ADDRESS"},{"prim":"SENDER"},[[{"prim":"COMPARE"},{"prim":"EQ"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"EXEC"},{"prim":"PAIR"}],[{"prim":"DROP"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]]}],"storage":{"bytes":"02baea6d6e4f1133f81dedd3f641296938e3996a7f"}}`,
			fields: fields{
				legacy:     true,
				statusCode: http.StatusOK,
			},
		}, {
			name:   "not legacy",
			script: `{"code":[{"prim":"parameter","args":[{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%do"]},{"prim":"unit","annots":["%default"]}]}]},{"prim":"storage","args":[{"prim":"key_hash"}]},{"prim":"code","args":[[[[{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]}]],{"prim":"IF_LEFT","args":[[{"prim":"PUSH","args":[{"prim":"mutez"},{"int":"0"}]},{"prim":"AMOUNT"},[[{"prim":"COMPARE"},{"prim":"EQ"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],[{"prim":"DIP","args":[[{"prim":"DUP"}]]},{"prim":"SWAP"}],{"prim":"IMPLICIT_ACCOUNT"},{"prim":"ADDRESS"},{"prim":"SENDER"},[[{"prim":"COMPARE"},{"prim":"EQ"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"EXEC"},{"prim":"PAIR"}],[{"prim":"DROP"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]]}],"storage":{"bytes":"02baea6d6e4f1133f81dedd3f641296938e3996a7f"}}`,
			fields: fields{
				legacy:     false,
				statusCode: http.StatusOK,
			},
		}, {
			name: "invalid status code",
			fields: fields{
				legacy:     false,
				statusCode: http.StatusNotFound,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := getHandler(tt.fields.statusCode, tt.script) // nolint
			rpc := newTestRPC(handler, tt.fields.legacy)
			got, err := rpc.GetScriptJSON("address", 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRPC.GetScriptJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := gjson.Parse(tt.script)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("NodeRPC.GetScriptJSON() = %v, want %v", got, want)
			}
		})
	}
}
