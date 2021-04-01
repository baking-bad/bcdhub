package main

import "testing"

func Test_parseID(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want int64
	}{
		{
			name: "Without quotes",
			data: []byte("1"),
			want: 1,
		}, {
			name: "With quotes",
			data: []byte(`"2"`),
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseID(tt.data)
			if err != nil {
				t.Errorf("parseID() error = %s", err.Error())
				return
			}
			if got != tt.want {
				t.Errorf("parseID() = %v, want %v", got, tt.want)
			}
		})
	}
}
