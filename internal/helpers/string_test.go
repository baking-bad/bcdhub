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

func TestIsIPFS(t *testing.T) {
	tests := []struct {
		name string
		hash string
		want bool
	}{
		{
			name: "QmSepFyoj44Xeok63ZVtSPehWg73dkGaUHAXyAwoeNyHj2",
			hash: "QmSepFyoj44Xeok63ZVtSPehWg73dkGaUHAXyAwoeNyHj2",
			want: true,
		}, {
			name: "QmX7Bhf8Gw7BDdyyGLPrsdfyf9pyqgAY4M3BzzSVwULxZy",
			hash: "QmX7Bhf8Gw7BDdyyGLPrsdfyf9pyqgAY4M3BzzSVwULxZy",
			want: true,
		}, {
			name: "QmaUx4RSLEiDrvKgb5NPtk1EHkcJHBQCWFcHQxTSep4VUw",
			hash: "QmaUx4RSLEiDrvKgb5NPtk1EHkcJHBQCWFcHQxTSep4VUw",
			want: true,
		}, {
			name: "QmX7Bhf8Gw7BDdyyGLPrsdfyf9pyqgAY4M3BzzSVwULxZ",
			hash: "QmX7Bhf8Gw7BDdyyGLPrsdfyf9pyqgAY4M3BzzSVwULxZ",
			want: false,
		}, {
			name: "Q1X7Bhf8Gw7BDdyyGLPrsdfyf9pyqgAY4M3BzzSVwULxZ",
			hash: "Q1X7Bhf8Gw7BDdyyGLPrsdfyf9pyqgAY4M3BzzSVwULxZ",
			want: false,
		}, {
			name: "undefined",
			hash: "undefined",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsIPFS(tt.hash); got != tt.want {
				t.Errorf("IsIPFS() = %v, want %v", got, tt.want)
			}
		})
	}
}
