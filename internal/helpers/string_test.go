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

func TestIsContract(t *testing.T) {
	tests := []struct {
		address string
		want    bool
	}{
		{
			address: "KT1A946hDgLGfFudWU7hzfnTdZK8TZyLRHeT",
			want:    true,
		}, {
			address: "tz3RNxKYP8Pt2LggVRKz5wQy66PH65kK2ZE2",
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.address, func(t *testing.T) {
			if got := IsContract(tt.address); got != tt.want {
				t.Errorf("IsContract() = %v, want %v", got, tt.want)
			}
		})
	}
}
