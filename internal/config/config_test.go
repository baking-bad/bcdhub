package config

import "testing"

func Test_expandEnv(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{
			name: "test 1",
			data: "${TEST}",
			want: "",
		}, {
			name: "test 2",
			data: "${TEST:-val}",
			want: "val",
		}, {
			name: "test 3",
			data: "${TEST:-val} ${TEST2:-}",
			want: "val ",
		}, {
			name: "test 4",
			data: "${TEST3:-127.0.0.1}",
			want: "127.0.0.1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := expandEnv(tt.data); got != tt.want {
				t.Errorf("expandEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
