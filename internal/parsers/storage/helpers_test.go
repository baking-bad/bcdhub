package storage

import (
	"reflect"
	"testing"

	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func Test_GetStrings(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    []string
		wantErr bool
	}{
		{
			name: "test 1",
			data: []byte(`{"bytes":"62616c6c732e74657a"}`),
			want: []string{
				"balls.tez",
			},
		}, {
			name: "test 2",
			data: []byte(`{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Some","args":[{"bytes":"0000c0ca282a775946b5ecbe02e5cf73e25f6b62b70c"}]},[]]},{"prim":"Pair","args":[{"prim":"Some","args":[{"bytes":"62616c6c732e74657a"}]},[]]}]},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"2"},{"bytes":"0000753f63893674b6d523f925f0d787bf9270b95c33"}]},{"prim":"Some","args":[{"int":"3223"}]}]}]}`),
			want: []string{
				"tz1dDQc4KsTHEFe3USc66Wti2pBatZ3UDbD4",
				"balls.tez",
				"tz1WKygtstVY96oyc6Rmk945dMf33LeihgWT",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetStrings(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStrings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setBigMapDiffsStrings(t *testing.T) {
	tests := []struct {
		name           string
		bmd            *bigmapdiff.BigMapDiff
		wantKeyStrings pq.StringArray
		wantErr        bool
	}{
		{
			name: "test 1",
			bmd: &bigmapdiff.BigMapDiff{
				Key: []byte(`{"bytes":"62616c6c732e74657a"}`),
			},
			wantKeyStrings: []string{
				"balls.tez",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setBigMapDiffsStrings(tt.bmd); (err != nil) != tt.wantErr {
				t.Errorf("setBigMapDiffsStrings() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.wantKeyStrings, tt.bmd.KeyStrings)
		})
	}
}
