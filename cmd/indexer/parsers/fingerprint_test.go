package parsers

import (
	"testing"

	"github.com/tidwall/gjson"
)

func Test_getCode(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    string
		wantErr bool
	}{
		{
			name:    "bool",
			args:    "bool",
			want:    "59",
			wantErr: false,
		}, {
			name:    "CMPGT",
			args:    "CMPGT",
			want:    "76",
			wantErr: false,
		}, {
			name:    "UNPPAAIR",
			args:    "UNPPAAIR",
			want:    "8e",
			wantErr: false,
		}, {
			name:    "PAPAPAIR",
			args:    "PAPAPAIR",
			want:    "8f",
			wantErr: false,
		}, {
			name:    "DIIIIP",
			args:    "DIIIIP",
			want:    "8d",
			wantErr: false,
		}, {
			name:    "DUUUUUP",
			args:    "DUUUUUP",
			want:    "8c",
			wantErr: false,
		}, {
			name:    "CADDDR",
			args:    "CADDDR",
			want:    "90",
			wantErr: false,
		}, {
			name:    "unknown",
			args:    "unknown",
			want:    "00",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getCode(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_skipPrim(t *testing.T) {
	type args struct {
		prim   string
		isCode bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "is code skip cast",
			args: args{
				prim:   "cast",
				isCode: true,
			},
			want: true,
		}, {
			name: "is code skip rename",
			args: args{
				prim:   "RENAME",
				isCode: true,
			},
			want: true,
		}, {
			name: "is code pass",
			args: args{
				prim:   "SET",
				isCode: true,
			},
			want: false,
		}, {
			name: "is code = false",
			args: args{
				prim:   "SET",
				isCode: false,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := skip(tt.args.prim, tt.args.isCode); got != tt.want {
				t.Errorf("skipPrim() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fingerprint(t *testing.T) {
	type args struct {
		script string
		isCode bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "simple prim",
			args: args{
				script: `{ "prim": "string" }`,
				isCode: false,
			},
			want:    "68",
			wantErr: false,
		}, {
			name: "simple prim 2",
			args: args{
				script: `{ "prim": "UNPAPAPAIR" }`,
				isCode: false,
			},
			want:    "8e",
			wantErr: false,
		}, {
			name: "code",
			args: args{
				script: `{ "prim": "code", "args":[{"prim": "CAST", "args":[{"prim": "string"}]}, { "prim": "string" }] }`,
				isCode: true,
			},
			want:    "0268",
			wantErr: false,
		}, {
			name: "parameter",
			args: args{
				script: `{ "prim": "parameter", "args":[{"prim": "or", "args":[{"prim": "string"}, { "string": "string" }]}]}`,
				isCode: false,
			},
			want:    "006868",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := gjson.Parse(tt.args.script)
			got, err := fingerprint(s, tt.args.isCode)
			if (err != nil) != tt.wantErr {
				t.Errorf("fingerprint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("fingerprint() = %s, want %s", got, tt.want)
			}
		})
	}
}
