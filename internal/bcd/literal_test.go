package bcd

import "testing"

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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsContract(tt.address); got != tt.want {
				t.Errorf("IsContract() = %v, want %v", got, tt.want)
			}
		})
	}
}
