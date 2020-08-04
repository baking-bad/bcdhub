package kinds

import (
	"reflect"
	"testing"
)

func TestCheckParameterForTags(t *testing.T) {
	tests := []struct {
		name      string
		parameter string
		want      []string
		wantErr   bool
	}{
		{
			name: "ViewAddress",
			parameter: `{
				"prim": "address"
			}`,
			want: []string{"view_address"},
		}, {
			name: "ViewNat",
			parameter: `{
				"prim": "nat"
			}`,
			want: []string{"view_nat"},
		}, {
			name: "ViewBalanceOf",
			parameter: `{
				"prim": "list",
				"args": [
					{
						"prim": "pair",
						"args": [
							{
								"prim": "pair",
								"args": [
									{
										"prim": "address"
									},
									{
										"prim": "nat"
									}
								]
							},
							{
								"prim": "nat"
							}
						]
					}
				]
			}`,
			want: []string{"view_balance_of"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckParameterForTags(tt.parameter)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckParameterForTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckParameterForTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
