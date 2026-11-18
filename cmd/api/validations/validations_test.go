package validations

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/stretchr/testify/require"
)

func TestIsBase58Check(t *testing.T) {
	tests := []struct {
		value   string
		isValid func(string) bool
		want    bool
	}{
		{"tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE", bcd.IsAddress, true},
		{"KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU", bcd.IsContract, true},
		// the last character changed: the shape is right, the checksum is not
		{"tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTF", bcd.IsAddress, false},
		{"KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoV", bcd.IsContract, false},
		{"KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU", bcd.IsSmartRollupHash, false},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			require.Equal(t, tt.want, isBase58Check(tt.value, tt.isValid))
		})
	}
}
