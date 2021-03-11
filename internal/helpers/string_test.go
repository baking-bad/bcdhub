package helpers

import (
	"testing"
)

func TestSlug(t *testing.T) {
	tests := []struct {
		alias string
		want  string
	}{
		{
			alias: "Aspen Coin",
			want:  "aspen-coin",
		}, {
			alias: "Trianon STO",
			want:  "trianon-sto",
		}, {
			alias: "Atomex",
			want:  "atomex",
		},
	}
	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			if got := Slug(tt.alias); got != tt.want {
				t.Errorf("Slug() = %v, want %v", got, tt.want)
			}
		})
	}
}
