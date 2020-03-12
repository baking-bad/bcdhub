package meta

import (
	"encoding/json"
	"testing"

	"github.com/tidwall/gjson"
)

func Test_getPath(t *testing.T) {
	type args struct {
		node string
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test default",
			args: args{
				node: `{"int": 3}`,
				path: "0",
			},
			want: "0",
		}, {
			name: "test default 2",
			args: args{
				node: `{"int": 3}`,
				path: "0/1",
			},
			want: "0/1",
		}, {
			name: "test left",
			args: args{
				node: `{"prim": "Left", "args":[{}]}`,
				path: "0",
			},
			want: "0/0",
		}, {
			name: "test left 2 ",
			args: args{
				node: `{"prim": "Left", "args":[{}]}`,
				path: "0/1",
			},
			want: "0/1/0",
		}, {
			name: "test right",
			args: args{
				node: `{"prim": "Right", "args":[{}]}`,
				path: "0",
			},
			want: "0/1",
		}, {
			name: "test right 2 ",
			args: args{
				node: `{"prim": "Right", "args":[{}]}`,
				path: "0/1",
			},
			want: "0/1/1",
		}, {
			name: "test left right",
			args: args{
				node: `{"prim": "Left", "args":[{"prim": "Right", "args":[{}]}]}`,
				path: "0",
			},
			want: "0/0/1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := gjson.Parse(tt.args.node)
			if got := getPath(node, tt.args.path); got != tt.want {
				t.Errorf("getPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGet(t *testing.T) {
	type args struct {
		node     string
		metadata string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "case 1",
			args: args{
				node:     `{"prim":"Right","args":[{"prim":"None"}]}`,
				metadata: `{"0":{"prim":"or","args":["0/0","0/1/o"],"type":"namedunion"},"0/0":{"fieldname":"Action","prim":"pair","args":["0/0/0","0/0/1"],"type":"namedtuple","name":"Action"},"0/0/0":{"fieldname":"action_input","prim":"or","args":["0/0/0/0/0","0/0/0/0/1/o","0/0/0/1/0","0/0/0/1/1/o"],"type":"namedunion","name":"action_input"},"0/0/0/0":{"prim":"or","type":"or"},"0/0/0/0/0":{"fieldname":"Transfer","prim":"pair","args":["0/0/0/0/0/0","0/0/0/0/0/1"],"type":"namedtuple","name":"Transfer"},"0/0/0/0/0/0":{"fieldname":"dest","prim":"contract","parameter":"{\"prim\":\"unit\"}","type":"contract","name":"dest"},"0/0/0/0/0/1":{"fieldname":"transfer_amount","prim":"mutez","type":"mutez","name":"transfer_amount"},"0/0/0/0/1":{"fieldname":"Set_pour","prim":"option","type":"option"},"0/0/0/0/1/o":{"prim":"pair","args":["0/0/0/0/1/o/0","0/0/0/0/1/o/1"],"type":"namedtuple","name":"Set_pour"},"0/0/0/0/1/o/0":{"fieldname":"pour_dest","prim":"contract","parameter":"{\"prim\":\"unit\"}","type":"contract","name":"pour_dest"},"0/0/0/0/1/o/1":{"fieldname":"pour_authorizer","prim":"key","type":"key","name":"pour_authorizer"},"0/0/0/1":{"prim":"or","type":"or"},"0/0/0/1/0":{"fieldname":"Set_keys","prim":"pair","args":["0/0/0/1/0/0","0/0/0/1/0/1"],"type":"namedtuple","name":"Set_keys"},"0/0/0/1/0/0":{"fieldname":"key_groups","prim":"list","type":"list","name":"key_groups"},"0/0/0/1/0/0/l":{"prim":"pair","args":["0/0/0/1/0/0/l/0","0/0/0/1/0/0/l/1"],"type":"namedtuple"},"0/0/0/1/0/0/l/0":{"fieldname":"signatories","prim":"list","type":"list","name":"signatories"},"0/0/0/1/0/0/l/0/l":{"prim":"key","type":"key"},"0/0/0/1/0/0/l/1":{"fieldname":"group_threshold","prim":"nat","type":"nat","name":"group_threshold"},"0/0/0/1/0/1":{"fieldname":"overall_threshold","prim":"nat","type":"nat","name":"overall_threshold"},"0/0/0/1/1":{"fieldname":"Set_delegate","prim":"option","type":"option"},"0/0/0/1/1/o":{"fieldname":"new_delegate","prim":"key_hash","type":"key_hash","name":"Set_delegate"},"0/0/1":{"fieldname":"signatures","prim":"list","type":"list","name":"signatures"},"0/0/1/l":{"prim":"list","type":"list"},"0/0/1/l/l":{"prim":"option","type":"option"},"0/0/1/l/l/o":{"prim":"signature","type":"signature"},"0/1":{"prim":"option","type":"option"},"0/1/o":{"fieldname":"Pour","prim":"pair","args":["0/1/o/0","0/1/o/1"],"type":"namedtuple","name":"Pour"},"0/1/o/0":{"fieldname":"pour_auth","prim":"signature","type":"signature","name":"pour_auth"},"0/1/o/1":{"fieldname":"pour_amount","prim":"mutez","type":"mutez","name":"pour_amount"}}`,
			},
			want:    "Pour",
			wantErr: false,
		}, {
			name: "case 2",
			args: args{
				node:     `{"prim":"Right","args":[{"prim":"Some","args":[{"prim":"Pair","args":[{"string":"edsigterpkPuRtc2hyS2yvNysmsDb8HBXVJSdiQd7nYwYUZCKHVEBUAwitrk3waQY6LJsyJRNp6NZFi7JmB2T6RAsMbbm717n9D"},{"int":"199041301565"}]}]}]}`,
				metadata: `{"0":{"prim":"or","args":["0/0","0/1/o"],"type":"namedunion"},"0/0":{"fieldname":"Action","prim":"pair","args":["0/0/0","0/0/1"],"type":"namedtuple","name":"Action"},"0/0/0":{"fieldname":"action_input","prim":"or","args":["0/0/0/0/0","0/0/0/0/1/o","0/0/0/1/0","0/0/0/1/1/o"],"type":"namedunion","name":"action_input"},"0/0/0/0":{"prim":"or","type":"or"},"0/0/0/0/0":{"fieldname":"Transfer","prim":"pair","args":["0/0/0/0/0/0","0/0/0/0/0/1"],"type":"namedtuple","name":"Transfer"},"0/0/0/0/0/0":{"fieldname":"dest","prim":"contract","parameter":"{\"prim\":\"unit\"}","type":"contract","name":"dest"},"0/0/0/0/0/1":{"fieldname":"transfer_amount","prim":"mutez","type":"mutez","name":"transfer_amount"},"0/0/0/0/1":{"fieldname":"Set_pour","prim":"option","type":"option"},"0/0/0/0/1/o":{"prim":"pair","args":["0/0/0/0/1/o/0","0/0/0/0/1/o/1"],"type":"namedtuple","name":"Set_pour"},"0/0/0/0/1/o/0":{"fieldname":"pour_dest","prim":"contract","parameter":"{\"prim\":\"unit\"}","type":"contract","name":"pour_dest"},"0/0/0/0/1/o/1":{"fieldname":"pour_authorizer","prim":"key","type":"key","name":"pour_authorizer"},"0/0/0/1":{"prim":"or","type":"or"},"0/0/0/1/0":{"fieldname":"Set_keys","prim":"pair","args":["0/0/0/1/0/0","0/0/0/1/0/1"],"type":"namedtuple","name":"Set_keys"},"0/0/0/1/0/0":{"fieldname":"key_groups","prim":"list","type":"list","name":"key_groups"},"0/0/0/1/0/0/l":{"prim":"pair","args":["0/0/0/1/0/0/l/0","0/0/0/1/0/0/l/1"],"type":"namedtuple"},"0/0/0/1/0/0/l/0":{"fieldname":"signatories","prim":"list","type":"list","name":"signatories"},"0/0/0/1/0/0/l/0/l":{"prim":"key","type":"key"},"0/0/0/1/0/0/l/1":{"fieldname":"group_threshold","prim":"nat","type":"nat","name":"group_threshold"},"0/0/0/1/0/1":{"fieldname":"overall_threshold","prim":"nat","type":"nat","name":"overall_threshold"},"0/0/0/1/1":{"fieldname":"Set_delegate","prim":"option","type":"option"},"0/0/0/1/1/o":{"fieldname":"new_delegate","prim":"key_hash","type":"key_hash","name":"Set_delegate"},"0/0/1":{"fieldname":"signatures","prim":"list","type":"list","name":"signatures"},"0/0/1/l":{"prim":"list","type":"list"},"0/0/1/l/l":{"prim":"option","type":"option"},"0/0/1/l/l/o":{"prim":"signature","type":"signature"},"0/1":{"prim":"option","type":"option"},"0/1/o":{"fieldname":"Pour","prim":"pair","args":["0/1/o/0","0/1/o/1"],"type":"namedtuple","name":"Pour"},"0/1/o/0":{"fieldname":"pour_auth","prim":"signature","type":"signature","name":"pour_auth"},"0/1/o/1":{"fieldname":"pour_amount","prim":"mutez","type":"mutez","name":"pour_amount"}}`,
			},
			want:    "Pour",
			wantErr: false,
		}, {
			name: "case 3",
			args: args{
				node:     `{"prim":"Left","args":[{"prim":"Pair","args":[{"string":"KT1R3uoZ6W1ZxEwzqtv75Ro7DhVY6UAcxuK2"},{"prim":"Pair","args":[{"string":"Aliases Contract"},{"prim":"None"}]}]}]}`,
				metadata: `{"0":{"prim":"or","args":["0/0","0/1/0","0/1/1/0","0/1/1/1"],"type":"union"},"0/0":{"prim":"pair","args":["0/0/0","0/0/1/0","0/0/1/1/o"],"type":"tuple"},"0/0/0":{"prim":"address","type":"address"},"0/0/1":{"prim":"pair","type":"pair"},"0/0/1/0":{"prim":"string","type":"string"},"0/0/1/1":{"prim":"option","type":"option"},"0/0/1/1/o":{"prim":"bytes","type":"bytes"},"0/1":{"prim":"or","type":"or"},"0/1/0":{"prim":"mutez","type":"mutez"},"0/1/1":{"prim":"or","type":"or"},"0/1/1/0":{"prim":"pair","args":["0/1/1/0/0","0/1/1/0/1"],"type":"namedtuple"},"0/1/1/0/0":{"prim":"address","type":"address","name":"address"},"0/1/1/0/1":{"prim":"bool","type":"bool"},"0/1/1/1":{"prim":"lambda","parameter":"{\"prim\":\"pair\",\"args\":[{\"prim\":\"address\"},{\"prim\":\"pair\",\"args\":[{\"prim\":\"string\"},{\"prim\":\"option\",\"args\":[{\"prim\":\"bytes\"}]}]}]}","type":"lambda"}}`,
			},
			want:    "entrypoint_0",
			wantErr: false,
		}, {
			name: "case 4",
			args: args{
				node:     `{"entrypoint":"main","value":{"prim":"Unit"}}`,
				metadata: `{"0":{"prim":"or","args":["0/0","0/1"],"type":"namedunion"},"0/0":{"fieldname":"main","prim":"unit","type":"unit","name":"main"},"0/1":{"fieldname":"set_delegate","prim":"key_hash","type":"key_hash","name":"set_delegate"}}`,
			},
			want:    "main",
			wantErr: false,
		}, {
			name: "case 5",
			args: args{
				node:     `{"entrypoint":"safeEntrypoints","value":{"prim":"Right","args":[{"prim":"Left","args":[{"prim":"Left","args":[{"prim":"Left","args":[{"prim":"Pair","args":[{"bytes":"0000a5c883b732a80872cd1ef9d6f33702caa860fadc"},{"int":"100500"}]}]}]}]}]}}`,
				metadata: `{"0":{"prim":"or","args":["0/0/0/0","0/0/0/1","0/0/1/0","0/0/1/1/0","0/0/1/1/1","0/1/0/0","0/1/0/1/0","0/1/0/1/1","0/1/1/0","0/1/1/1/0","0/1/1/1/1/0/0/0/0","0/1/1/1/1/0/0/0/1","0/1/1/1/1/0/0/1/0","0/1/1/1/1/0/0/1/1","0/1/1/1/1/0/1/0/0","0/1/1/1/1/0/1/0/1","0/1/1/1/1/0/1/1/0","0/1/1/1/1/0/1/1/1","0/1/1/1/1/1/0/0/0","0/1/1/1/1/1/0/0/1","0/1/1/1/1/1/0/1/0","0/1/1/1/1/1/0/1/1","0/1/1/1/1/1/1/0/0","0/1/1/1/1/1/1/0/1","0/1/1/1/1/1/1/1/0","0/1/1/1/1/1/1/1/1/0","0/1/1/1/1/1/1/1/1/1"],"type":"namedunion"},"0/0":{"prim":"or","type":"or"},"0/0/0":{"prim":"or","type":"or"},"0/0/0/0":{"fieldname":"getVersion","prim":"pair","args":["0/0/0/0/0","0/0/0/0/1"],"type":"tuple","name":"getVersion"},"0/0/0/0/0":{"prim":"unit","type":"unit"},"0/0/0/0/1":{"prim":"contract","parameter":"{\"prim\":\"nat\"}","type":"contract"},"0/0/0/1":{"fieldname":"getAllowance","prim":"pair","args":["0/0/0/1/0/0","0/0/0/1/0/1","0/0/0/1/1"],"type":"namedtuple","name":"getAllowance"},"0/0/0/1/0":{"prim":"pair","type":"pair"},"0/0/0/1/0/0":{"typename":"owner","prim":"address","type":"address","name":"owner"},"0/0/0/1/0/1":{"typename":"spender","prim":"address","type":"address","name":"spender"},"0/0/0/1/1":{"prim":"contract","parameter":"{\"prim\":\"nat\"}","type":"contract"},"0/0/1":{"prim":"or","type":"or"},"0/0/1/0":{"fieldname":"getBalance","prim":"pair","args":["0/0/1/0/0","0/0/1/0/1"],"type":"namedtuple","name":"getBalance"},"0/0/1/0/0":{"typename":"owner","prim":"address","type":"address","name":"owner"},"0/0/1/0/1":{"prim":"contract","parameter":"{\"prim\":\"nat\"}","type":"contract"},"0/0/1/1":{"prim":"or","type":"or"},"0/0/1/1/0":{"fieldname":"getTotalSupply","prim":"pair","args":["0/0/1/1/0/0","0/0/1/1/0/1"],"type":"tuple","name":"getTotalSupply"},"0/0/1/1/0/0":{"prim":"unit","type":"unit"},"0/0/1/1/0/1":{"prim":"contract","parameter":"{\"prim\":\"nat\"}","type":"contract"},"0/0/1/1/1":{"fieldname":"getTotalMinted","prim":"pair","args":["0/0/1/1/1/0","0/0/1/1/1/1"],"type":"tuple","name":"getTotalMinted"},"0/0/1/1/1/0":{"prim":"unit","type":"unit"},"0/0/1/1/1/1":{"prim":"contract","parameter":"{\"prim\":\"nat\"}","type":"contract"},"0/1":{"prim":"or","type":"or"},"0/1/0":{"prim":"or","type":"or"},"0/1/0/0":{"fieldname":"getTotalBurned","prim":"pair","args":["0/1/0/0/0","0/1/0/0/1"],"type":"tuple","name":"getTotalBurned"},"0/1/0/0/0":{"prim":"unit","type":"unit"},"0/1/0/0/1":{"prim":"contract","parameter":"{\"prim\":\"nat\"}","type":"contract"},"0/1/0/1":{"prim":"or","type":"or"},"0/1/0/1/0":{"fieldname":"getOwner","prim":"pair","args":["0/1/0/1/0/0","0/1/0/1/0/1"],"type":"tuple","name":"getOwner"},"0/1/0/1/0/0":{"prim":"unit","type":"unit"},"0/1/0/1/0/1":{"prim":"contract","parameter":"{\"prim\":\"address\"}","type":"contract"},"0/1/0/1/1":{"fieldname":"getTokenName","prim":"pair","args":["0/1/0/1/1/0","0/1/0/1/1/1"],"type":"tuple","name":"getTokenName"},"0/1/0/1/1/0":{"prim":"unit","type":"unit"},"0/1/0/1/1/1":{"prim":"contract","parameter":"{\"prim\":\"string\"}","type":"contract"},"0/1/1":{"prim":"or","type":"or"},"0/1/1/0":{"fieldname":"getTokenCode","prim":"pair","args":["0/1/1/0/0","0/1/1/0/1"],"type":"tuple","name":"getTokenCode"},"0/1/1/0/0":{"prim":"unit","type":"unit"},"0/1/1/0/1":{"prim":"contract","parameter":"{\"prim\":\"string\"}","type":"contract"},"0/1/1/1":{"prim":"or","type":"or"},"0/1/1/1/0":{"fieldname":"getRedeemAddress","prim":"pair","args":["0/1/1/1/0/0","0/1/1/1/0/1"],"type":"tuple","name":"getRedeemAddress"},"0/1/1/1/0/0":{"prim":"unit","type":"unit"},"0/1/1/1/0/1":{"prim":"contract","parameter":"{\"prim\":\"address\"}","type":"contract"},"0/1/1/1/1":{"fieldname":"safeEntrypoints","prim":"or","type":"or"},"0/1/1/1/1/0":{"prim":"or","type":"or"},"0/1/1/1/1/0/0":{"prim":"or","type":"or"},"0/1/1/1/1/0/0/0":{"prim":"or","type":"or"},"0/1/1/1/1/0/0/0/0":{"fieldname":"run","prim":"pair","args":["0/1/1/1/1/0/0/0/0/0","0/1/1/1/1/0/0/0/0/1"],"type":"tuple","name":"run"},"0/1/1/1/1/0/0/0/0/0":{"prim":"string","type":"string"},"0/1/1/1/1/0/0/0/0/1":{"prim":"bytes","type":"bytes"},"0/1/1/1/1/0/0/0/1":{"fieldname":"upgrade","prim":"pair","args":["0/1/1/1/1/0/0/0/1/0/0","0/1/1/1/1/0/0/0/1/0/1","0/1/1/1/1/0/0/0/1/1/0","0/1/1/1/1/0/0/0/1/1/1/0/o","0/1/1/1/1/0/0/0/1/1/1/1/o"],"type":"tuple","name":"upgrade"},"0/1/1/1/1/0/0/0/1/0":{"prim":"pair","type":"pair"},"0/1/1/1/1/0/0/0/1/0/0":{"typename":"currentVersion","prim":"nat","type":"nat"},"0/1/1/1/1/0/0/0/1/0/1":{"typename":"newVersion","prim":"nat","type":"nat"},"0/1/1/1/1/0/0/0/1/1":{"prim":"pair","type":"pair"},"0/1/1/1/1/0/0/0/1/1/0":{"typename":"migrationScript","prim":"lambda","parameter":"{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}","type":"lambda"},"0/1/1/1/1/0/0/0/1/1/1":{"prim":"pair","type":"pair"},"0/1/1/1/1/0/0/0/1/1/1/0":{"typename":"newCode","prim":"option","type":"option"},"0/1/1/1/1/0/0/0/1/1/1/0/o":{"prim":"lambda","parameter":"{\"prim\":\"pair\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"string\"},{\"prim\":\"bytes\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}","type":"lambda"},"0/1/1/1/1/0/0/0/1/1/1/1":{"typename":"newPermCode","prim":"option","type":"option"},"0/1/1/1/1/0/0/0/1/1/1/1/o":{"prim":"lambda","parameter":"{\"prim\":\"pair\",\"args\":[{\"prim\":\"unit\"},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}","type":"lambda"},"0/1/1/1/1/0/0/1":{"prim":"or","type":"or"},"0/1/1/1/1/0/0/1/0":{"fieldname":"epwBeginUpgrade","prim":"pair","args":["0/1/1/1/1/0/0/1/0/0","0/1/1/1/1/0/0/1/0/1"],"type":"namedtuple","name":"epwBeginUpgrade"},"0/1/1/1/1/0/0/1/0/0":{"typename":"current","prim":"nat","type":"nat","name":"current"},"0/1/1/1/1/0/0/1/0/1":{"typename":"new","prim":"nat","type":"nat","name":"new"},"0/1/1/1/1/0/0/1/1":{"typename":"migrationscript","fieldname":"epwApplyMigration","prim":"lambda","parameter":"{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}","type":"lambda","name":"epwApplyMigration"},"0/1/1/1/1/0/1":{"prim":"or","type":"or"},"0/1/1/1/1/0/1/0":{"prim":"or","type":"or"},"0/1/1/1/1/0/1/0/0":{"typename":"contractcode","fieldname":"epwSetCode","prim":"lambda","parameter":"{\"prim\":\"pair\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"string\"},{\"prim\":\"bytes\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}","type":"lambda","name":"epwSetCode"},"0/1/1/1/1/0/1/0/1":{"fieldname":"epwFinishUpgrade","prim":"unit","type":"unit","name":"epwFinishUpgrade"},"0/1/1/1/1/0/1/1":{"prim":"or","type":"or"},"0/1/1/1/1/0/1/1/0":{"fieldname":"transfer","prim":"pair","args":["0/1/1/1/1/0/1/1/0/0","0/1/1/1/1/0/1/1/0/1/0","0/1/1/1/1/0/1/1/0/1/1"],"type":"namedtuple","name":"transfer"},"0/1/1/1/1/0/1/1/0/0":{"typename":"from","prim":"address","type":"address","name":"from"},"0/1/1/1/1/0/1/1/0/1":{"prim":"pair","type":"pair"},"0/1/1/1/1/0/1/1/0/1/0":{"typename":"to","prim":"address","type":"address","name":"to"},"0/1/1/1/1/0/1/1/0/1/1":{"typename":"value","prim":"nat","type":"nat","name":"value"},"0/1/1/1/1/0/1/1/1":{"fieldname":"approve","prim":"pair","args":["0/1/1/1/1/0/1/1/1/0","0/1/1/1/1/0/1/1/1/1"],"type":"namedtuple","name":"approve"},"0/1/1/1/1/0/1/1/1/0":{"typename":"spender","prim":"address","type":"address","name":"spender"},"0/1/1/1/1/0/1/1/1/1":{"typename":"value","prim":"nat","type":"nat","name":"value"},"0/1/1/1/1/1":{"prim":"or","type":"or"},"0/1/1/1/1/1/0":{"prim":"or","type":"or"},"0/1/1/1/1/1/0/0":{"prim":"or","type":"or"},"0/1/1/1/1/1/0/0/0":{"fieldname":"mint","prim":"pair","args":["0/1/1/1/1/1/0/0/0/0","0/1/1/1/1/1/0/0/0/1"],"type":"namedtuple","name":"mint"},"0/1/1/1/1/1/0/0/0/0":{"typename":"to","prim":"address","type":"address","name":"to"},"0/1/1/1/1/1/0/0/0/1":{"typename":"value","prim":"nat","type":"nat","name":"value"},"0/1/1/1/1/1/0/0/1":{"typename":"value","fieldname":"burn","prim":"nat","type":"nat","name":"burn"},"0/1/1/1/1/1/0/1":{"prim":"or","type":"or"},"0/1/1/1/1/1/0/1/0":{"typename":"operator","fieldname":"addOperator","prim":"address","type":"address","name":"addOperator"},"0/1/1/1/1/1/0/1/1":{"typename":"operator","fieldname":"removeOperator","prim":"address","type":"address","name":"removeOperator"},"0/1/1/1/1/1/1":{"prim":"or","type":"or"},"0/1/1/1/1/1/1/0":{"prim":"or","type":"or"},"0/1/1/1/1/1/1/0/0":{"typename":"redeem","fieldname":"setRedeemAddress","prim":"address","type":"address","name":"setRedeemAddress"},"0/1/1/1/1/1/1/0/1":{"fieldname":"pause","prim":"unit","type":"unit","name":"pause"},"0/1/1/1/1/1/1/1":{"prim":"or","type":"or"},"0/1/1/1/1/1/1/1/0":{"fieldname":"unpause","prim":"unit","type":"unit","name":"unpause"},"0/1/1/1/1/1/1/1/1":{"prim":"or","type":"or"},"0/1/1/1/1/1/1/1/1/0":{"typename":"newOwner","fieldname":"transferOwnership","prim":"address","type":"address","name":"transferOwnership"},"0/1/1/1/1/1/1/1/1/1":{"fieldname":"acceptOwnership","prim":"unit","type":"unit","name":"acceptOwnership"}}`,
			},
			want:    "mint",
			wantErr: false,
		}, {
			name: "case 6",
			args: args{
				node:     `{"entrypoint":"default","value":{"prim":"Left","args":[{"prim":"Left","args":[{"prim":"Left","args":[{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Some","args":[{"string":"KT1DSD5VoycG6TwpcYQMGES43rUvJxkAP31P"}]}]}]}]}]}]}}`,
				metadata: `{"0":{"prim":"or","args":["0/0/0/0/0/0","0/0/0/0/0/1/o","0/0/0/0/1","0/0/0/1","0/0/1","0/1"],"type":"namedunion"},"0/0":{"prim":"or","type":"or"},"0/0/0":{"prim":"or","type":"or"},"0/0/0/0":{"prim":"or","type":"or"},"0/0/0/0/0":{"prim":"or","type":"or"},"0/0/0/0/0/0":{"fieldname":"collateralize","prim":"mutez","type":"mutez","name":"collateralize"},"0/0/0/0/0/1":{"fieldname":"delegate","prim":"option","type":"option"},"0/0/0/0/0/1/o":{"prim":"key_hash","type":"key_hash","name":"delegate"},"0/0/0/0/1":{"fieldname":"deposit","prim":"pair","args":["0/0/0/0/1/0","0/0/0/0/1/1"],"type":"namedtuple","name":"deposit"},"0/0/0/0/1/0":{"fieldname":"duration","prim":"int","type":"int","name":"duration"},"0/0/0/0/1/1":{"fieldname":"rate","prim":"nat","type":"nat","name":"rate"},"0/0/0/1":{"fieldname":"setOffer","prim":"pair","args":["0/0/0/1/0","0/0/0/1/1"],"type":"namedtuple","name":"setOffer"},"0/0/0/1/0":{"fieldname":"duration","prim":"int","type":"int","name":"duration"},"0/0/0/1/1":{"fieldname":"rate","prim":"nat","type":"nat","name":"rate"},"0/0/1":{"fieldname":"uncollateralize","prim":"mutez","type":"mutez","name":"uncollateralize"},"0/1":{"fieldname":"withdraw","prim":"unit","type":"unit","name":"withdraw"}}`,
			},
			want:    "delegate",
			wantErr: false,
		}, {
			name: "case 7",
			args: args{
				node:     `{ "prim": "Left", "args": [ { "prim": "Some", "args": [ { "bytes": "0161335f2d0bb57d4b0a94552214104639cb955df500"}]}]}`,
				metadata: `{"0":{"typename":"_entries","prim":"or","args":["0/0/o","0/1/0","0/1/1/0","0/1/1/1/0","0/1/1/1/1"],"type":"namedunion"},"0/0":{"fieldname":"_Liq_entry_buy_for","prim":"option","type":"option"},"0/0/o":{"prim":"address","type":"address","name":"buy_for"},"0/1":{"prim":"or","type":"or"},"0/1/0":{"typename":"sell_request","fieldname":"_Liq_entry_sell_for","prim":"pair","args":["0/1/0/0/o","0/1/0/1/o"],"type":"namedtuple","name":"sell_for"},"0/1/0/0":{"fieldname":"buyer","prim":"option","type":"option"},"0/1/0/0/o":{"prim":"address","type":"address","name":"buyer"},"0/1/0/1":{"fieldname":"tokens","prim":"option","type":"option"},"0/1/0/1/o":{"prim":"mutez","type":"mutez","name":"tokens"},"0/1/1":{"prim":"or","type":"or"},"0/1/1/0":{"fieldname":"_Liq_entry_set_target_supply","prim":"mutez","type":"mutez","name":"set_target_supply"},"0/1/1/1":{"prim":"or","type":"or"},"0/1/1/1/0":{"typename":"sell_request","fieldname":"_Liq_entry_finalize_sale","prim":"pair","args":["0/1/1/1/0/0/o","0/1/1/1/0/1/o"],"type":"namedtuple","name":"finalize_sale"},"0/1/1/1/0/0":{"fieldname":"buyer","prim":"option","type":"option"},"0/1/1/1/0/0/o":{"prim":"address","type":"address","name":"buyer"},"0/1/1/1/0/1":{"fieldname":"tokens","prim":"option","type":"option"},"0/1/1/1/0/1/o":{"prim":"mutez","type":"mutez","name":"tokens"},"0/1/1/1/1":{"fieldname":"_Liq_entry_set_sell_adapter","prim":"address","type":"address","name":"set_sell_adapter"}}`,
			},
			want:    "buy_for",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := gjson.Parse(tt.args.node)
			var metadata Metadata
			if err := json.Unmarshal([]byte(tt.args.metadata), &metadata); err != nil {
				t.Errorf("Invalid metadata string: %v %s", err, tt.args.metadata)
				return
			}
			got, err := metadata.GetByPath(node)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
