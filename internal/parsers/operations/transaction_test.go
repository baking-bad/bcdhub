package operations

import (
	"reflect"
	"testing"

	"github.com/tidwall/gjson"
)

func Test_findByFieldName(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		data      string
		want      string
		wantErr   bool
	}{
		{
			name:      "default",
			fieldName: "default",
			data:      `{"prim": "unit"}`,
			want:      `{"prim": "unit"}`,
			wantErr:   true,
		}, {
			name:      "not found",
			fieldName: "test",
			data:      `{"prim": "unit"}`,
			want:      `{"prim": "unit"}`,
			wantErr:   true,
		}, {
			name:      "found main",
			fieldName: "main",
			data:      `{"prim":"or","args":[{"prim":"unit","annots":["%default"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%operation"]},{"prim":"pair","args":[{"prim":"nat"},{"prim":"list","args":[{"prim":"key"}]}],"annots":["%changeKeys"]}],"annots":[":action"]}],"annots":[":payload"]},{"prim":"list","args":[{"prim":"option","args":[{"prim":"signature"}]}]}],"annots":["%main"]}]}`,
			want:      `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%operation"]},{"prim":"pair","args":[{"prim":"nat"},{"prim":"list","args":[{"prim":"key"}]}],"annots":["%changeKeys"]}],"annots":[":action"]}],"annots":[":payload"]},{"prim":"list","args":[{"prim":"option","args":[{"prim":"signature"}]}]}],"annots":["%main"]}`,
		}, {
			name:      "found default",
			fieldName: "default",
			data:      `{"prim":"or","args":[{"prim":"unit","annots":["%default"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%operation"]},{"prim":"pair","args":[{"prim":"nat"},{"prim":"list","args":[{"prim":"key"}]}],"annots":["%changeKeys"]}],"annots":[":action"]}],"annots":[":payload"]},{"prim":"list","args":[{"prim":"option","args":[{"prim":"signature"}]}]}],"annots":["%main"]}]}`,
			want:      `{"prim":"unit","annots":["%default"]}`,
		}, {
			name:      "found default",
			fieldName: "default",
			data:      `{"args":[{"args":[{"args":[{"args":[{"prim":"address"},{"args":[{"prim":"address"},{"prim":"nat"}],"prim":"pair"}],"prim":"pair"},{"args":[{"args":[{"prim":"address"},{"args":[{"prim":"address"},{"args":[{"prim":"address"},{"prim":"nat"}],"prim":"pair"}],"prim":"pair"}],"prim":"pair"},{"args":[{"prim":"address"},{"prim":"nat"}],"prim":"pair"}],"prim":"or"}],"prim":"or"},{"args":[{"args":[{"prim":"address"},{"args":[{"prim":"address"},{"prim":"nat"}],"prim":"pair"}],"prim":"pair"},{"args":[{"args":[{"args":[{"prim":"address"},{"prim":"address"}],"prim":"pair"},{"args":[{"prim":"nat"}],"prim":"contract"}],"prim":"pair"},{"args":[{"prim":"address"},{"args":[{"prim":"nat"}],"prim":"contract"}],"prim":"pair"}],"prim":"or"}],"prim":"or"}],"prim":"or"},{"args":[{"args":[{"args":[{"prim":"unit"},{"args":[{"prim":"nat"}],"prim":"contract"}],"prim":"pair"},{"args":[{"prim":"bool"},{"prim":"address"}],"prim":"or"}],"prim":"or"},{"args":[{"args":[{"args":[{"prim":"unit"},{"args":[{"prim":"address"}],"prim":"contract"}],"prim":"pair"},{"args":[{"prim":"address"},{"prim":"nat"}],"prim":"pair"}],"prim":"or"},{"args":[{"args":[{"prim":"address"},{"prim":"nat"}],"prim":"pair"},{"prim":"address"}],"prim":"or"}],"prim":"or"}],"prim":"or"}],"prim":"or"}`,
			want:      `{"args":[{"args":[{"args":[{"args":[{"prim":"address"},{"args":[{"prim":"address"},{"prim":"nat"}],"prim":"pair"}],"prim":"pair"},{"args":[{"args":[{"prim":"address"},{"args":[{"prim":"address"},{"args":[{"prim":"address"},{"prim":"nat"}],"prim":"pair"}],"prim":"pair"}],"prim":"pair"},{"args":[{"prim":"address"},{"prim":"nat"}],"prim":"pair"}],"prim":"or"}],"prim":"or"},{"args":[{"args":[{"prim":"address"},{"args":[{"prim":"address"},{"prim":"nat"}],"prim":"pair"}],"prim":"pair"},{"args":[{"args":[{"args":[{"prim":"address"},{"prim":"address"}],"prim":"pair"},{"args":[{"prim":"nat"}],"prim":"contract"}],"prim":"pair"},{"args":[{"prim":"address"},{"args":[{"prim":"nat"}],"prim":"contract"}],"prim":"pair"}],"prim":"or"}],"prim":"or"}],"prim":"or"},{"args":[{"args":[{"args":[{"prim":"unit"},{"args":[{"prim":"nat"}],"prim":"contract"}],"prim":"pair"},{"args":[{"prim":"bool"},{"prim":"address"}],"prim":"or"}],"prim":"or"},{"args":[{"args":[{"args":[{"prim":"unit"},{"args":[{"prim":"address"}],"prim":"contract"}],"prim":"pair"},{"args":[{"prim":"address"},{"prim":"nat"}],"prim":"pair"}],"prim":"or"},{"args":[{"args":[{"prim":"address"},{"prim":"nat"}],"prim":"pair"},{"prim":"address"}],"prim":"or"}],"prim":"or"}],"prim":"or"}],"prim":"or"}`,
			wantErr:   true,
		}, {
			name:      "KT1TnwBxgK4ayHuxrti6KKkJpWBHXBYRCX6H",
			fieldName: "withdraw",
			data:      `{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"unit","annots":["%deposit"]},{"prim":"mutez","annots":["%reRoll"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":["%investor"]},{"prim":"address","annots":["%referrar"]}],"annots":["%register"]},{"prim":"or","args":[{"prim":"key_hash","annots":["%updateBaker"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%devFee"]},{"prim":"nat","annots":["%distributionPoolFee"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%referralFee"]},{"prim":"nat","annots":["%rewardPoolFee"]}]}],"annots":["%deposit"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%devFee"]},{"prim":"nat","annots":["%distributionPoolFee"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%referralFee"]},{"prim":"nat","annots":["%rewardPoolFee"]}]}],"annots":["%withdraw"]}],"annots":["%updateConfig"]}]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"address","annots":["%updateInvestorBaker"]},{"prim":"address","annots":["%updateOwner"]}]},{"prim":"or","args":[{"prim":"int","annots":["%updateRewardPoolReleaseDayCount"]},{"prim":"or","args":[{"prim":"mutez","annots":["%withdraw"]},{"prim":"unit","annots":["%withdrawDevFee"]}]}]}]}]}`,
			want:      `{"prim":"mutez","annots":["%withdraw"]}`,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := gjson.Parse(tt.data)
			got, err := findByFieldName(tt.fieldName, data)
			if (err != nil) != tt.wantErr {
				t.Errorf("findByFieldName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := gjson.Parse(tt.want)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("findByFieldName() = %v, want %v", got, want)
			}
		})
	}
}
