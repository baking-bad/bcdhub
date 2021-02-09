package ast

import "testing"

func TestAddress_Compare(t *testing.T) {
	tests := []struct {
		name       string
		first      string
		firstType  int
		second     string
		secondType int
		want       bool
		wantErr    bool
	}{
		{
			name:       "equal",
			first:      "KT1Hbwyp8D39d3681bG4FtZ1rE1uopVmU4wK",
			firstType:  valueTypeString,
			second:     "KT1Hbwyp8D39d3681bG4FtZ1rE1uopVmU4wK",
			secondType: valueTypeString,
			want:       true,
		}, {
			name:       "unequal",
			first:      "KT1Hbwyp8D39d3681bG4FtZ1rE1uopVmU4wK",
			firstType:  valueTypeString,
			second:     "KT1MjT5jseoujXvy1w2PjdaFXYo8jeh8k5S2",
			secondType: valueTypeString,
			want:       false,
		}, {
			name:       "equal",
			first:      "tz1eLWfccL46VAUjtyz9kEKgzuKnwyZH4rTA",
			firstType:  valueTypeString,
			second:     "0000cd1a410ffd5315ded34337f5f76edff48a13999a",
			secondType: valueTypeBytes,
			want:       true,
		}, {
			name:       "equal",
			first:      "0000cd1a410ffd5315ded34337f5f76edff48a13999a",
			firstType:  valueTypeBytes,
			second:     "tz1eLWfccL46VAUjtyz9kEKgzuKnwyZH4rTA",
			secondType: valueTypeString,
			want:       true,
		}, {
			name:       "equal",
			first:      "0000cd1a410ffd5315ded34337f5f76edff48a13999a",
			firstType:  valueTypeBytes,
			second:     "tz1eLWfccL46VAUjtyz9kEKgzuKnwyZH4rTA",
			secondType: valueTypeString,
			want:       true,
		}, {
			name:       "unequal",
			first:      "0000cd1a410ffd5315ded34337f5f76edff48a13999a",
			firstType:  valueTypeBytes,
			second:     "KT1DEkR3cErDAn6oH4jK8Z7n9a4oCXRZZwYa",
			secondType: valueTypeString,
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first := NewAddress(0)
			first.Value = tt.first
			first.valueType = tt.firstType

			second := NewAddress(0)
			second.Value = tt.second
			second.valueType = tt.secondType

			got, err := first.Compare(second)
			if (err != nil) != tt.wantErr {
				t.Errorf("Address.Compare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Address.Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}
