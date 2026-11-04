package bcd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsContract(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    bool
	}{
		{
			name:    "KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU",
			address: "KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU",
			want:    true,
		}, {
			name:    "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
			address: "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
			want:    false,
		}, {
			name:    "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
			address: "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
			want:    true,
		}, {
			name:    "txr1YNMEtkj5Vkqsbdmt7xaxBTMRZjzS96UAi",
			address: "txr1YNMEtkj5Vkqsbdmt7xaxBTMRZjzS96UAi",
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsContract(tt.address)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestIsAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    bool
	}{
		{
			name:    "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
			address: "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
			want:    true,
		}, {
			name:    "txr1YNMEtkj5Vkqsbdmt7xaxBTMRZjzS96UAi",
			address: "txr1YNMEtkj5Vkqsbdmt7xaxBTMRZjzS96UAi",
			want:    true,
		}, {
			name:    "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
			address: "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
			want:    true,
		}, {
			name:    "txr1YNMEtkj5Vkqsbdmt7xaxBTMRZjzS96UA",
			address: "txr1YNMEtkj5Vkqsbdmt7xaxBTMRZjzS96UA",
			want:    false,
		}, {
			name:    "sr1J1ECygUgzE7urU3Ayr5HZaty83hpjbs28",
			address: "sr1J1ECygUgzE7urU3Ayr5HZaty83hpjbs28",
			want:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAddress(tt.address)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestIsBakerHash(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{
			name: "SG1d1wsgMKvSstzZQ8L4WoskCesdWGzVt5k4",
			str:  "SG1d1wsgMKvSstzZQ8L4WoskCesdWGzVt5k4",
			want: true,
		}, {
			name: "SG1d1wsgMKvSstzZQ8L4WoskCesdWGzVt5k",
			str:  "SG1d1wsgMKvSstzZQ8L4WoskCesdWGzVt5k",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBakerHash(tt.str)
			require.Equal(t, tt.want, got)
		})
	}
}

// Validators used to match a substring, so text around a valid-looking
// address passed validation and reached node RPC URL paths.
func TestValidators_MatchWholeInput(t *testing.T) {
	const (
		tz1      = "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE"
		kt1      = "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr"
		sr1      = "sr1J1ECygUgzE7urU3Ayr5HZaty83hpjbs28"
		baker    = "SG1jfZeHRzeWAM1T4zrwunEyUpwWc82D4tbv"
		withZero = "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZS0" // 0 is not in base58
	)
	validators := map[string]func(string) bool{
		"IsAddress":         IsAddress,
		"IsContract":        IsContract,
		"IsBakerHash":       IsBakerHash,
		"IsSmartRollupHash": IsSmartRollupHash,
	}
	valid := map[string]string{
		"IsAddress":         tz1,
		"IsContract":        kt1,
		"IsBakerHash":       baker,
		"IsSmartRollupHash": sr1,
	}
	for name, validate := range validators {
		t.Run(name, func(t *testing.T) {
			value := valid[name]
			require.True(t, validate(value))
			for _, input := range []string{
				value + "/../../../chains/main/blocks/head/context/contracts/" + kt1,
				"../" + value,
				value + "\nforged log line",
				value + "?x=1",
				" " + value,
				value + "1",
				value[:len(value)-1],
			} {
				require.False(t, validate(input), input)
			}
		})
	}

	require.False(t, IsContract(withZero))
	require.False(t, IsAddress(withZero))
	require.False(t, IsAddress("tz9dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE"))
	require.False(t, IsSmartRollupHash("srXJ1ECygUgzE7urU3Ayr5HZaty83hpjbs28"))
}

func TestFindContract(t *testing.T) {
	const kt1 = "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr"
	require.Equal(t, kt1, FindContract("Error in contract "+kt1+": script failed"))
	require.Empty(t, FindContract("no contract here"))
}
