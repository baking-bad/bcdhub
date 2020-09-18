package main

import "testing"

func Test_parseID(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "Without quotes",
			data: []byte("test"),
			want: "test",
		}, {
			name: "With quotes",
			data: []byte(`"test"`),
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseID(tt.data); got != tt.want {
				t.Errorf("parseID() = %v, want %v", got, tt.want)
			}
		})
	}
}
