package ast

import (
	"math"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/stretchr/testify/assert"
)

func TestTimestamp_ParseValue(t *testing.T) {
	tests := []struct {
		name    string
		ts      string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "test: zero timestamp",
			ts:      "0",
			want:    time.Unix(0, 0).UTC(),
			wantErr: false,
		}, {
			name:    "test: timestamp overflow",
			ts:      "11111111111111111111111111",
			want:    time.Unix(math.MaxInt64, 0).UTC(),
			wantErr: false,
		}, {
			name:    "test: timestamp",
			ts:      "1624101720",
			want:    time.Unix(1624101720, 0).UTC(),
			wantErr: false,
		}, {
			name:    "test: timestamp milliseconds",
			ts:      "1624101720000",
			want:    time.Unix(1624101720, 0).UTC(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &Timestamp{
				Default: Default{},
			}
			node := &base.Node{
				StringValue: &tt.ts,
			}
			if err := ts.ParseValue(node); (err != nil) != tt.wantErr {
				t.Errorf("Timestamp.ParseValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, ts.Value)
		})
	}
}
