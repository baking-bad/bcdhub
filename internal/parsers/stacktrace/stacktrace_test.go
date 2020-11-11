package stacktrace

import (
	"testing"
)

func setInt64(x int64) *int64 {
	return &x
}

func Test_GetID(t *testing.T) {
	tests := []struct {
		name string
		sti  *Item
		want int64
	}{
		{
			name: "test 1",
			sti: &Item{
				contentIndex: 1,
			},
			want: 1000,
		}, {
			name: "test 2",
			sti:  &Item{},
			want: 0,
		}, {
			name: "test 3",
			sti: &Item{
				contentIndex: 3,
				nonce:        setInt64(2),
			},
			want: 3003,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sti.GetID(); got != tt.want {
				t.Errorf("getStackTraceID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStackTraceItem_gtNonce(t *testing.T) {
	type fields struct {
		nonce *int64
	}
	tests := []struct {
		name   string
		fields fields
		nonce  *int64
		want   bool
	}{
		{
			name: "test 1",
			fields: fields{
				nonce: nil,
			},
			nonce: setInt64(1),
			want:  true,
		}, {
			name: "test 2",
			fields: fields{
				nonce: nil,
			},
			nonce: nil,
			want:  false,
		}, {
			name: "test 3",
			fields: fields{
				nonce: setInt64(1),
			},
			nonce: nil,
			want:  false,
		}, {
			name: "test 4",
			fields: fields{
				nonce: setInt64(1),
			},
			nonce: setInt64(1),
			want:  false,
		}, {
			name: "test 5",
			fields: fields{
				nonce: setInt64(2),
			},
			nonce: setInt64(1),
			want:  false,
		}, {
			name: "test 6",
			fields: fields{
				nonce: setInt64(2),
			},
			nonce: setInt64(3),
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sti := &Item{
				nonce: tt.fields.nonce,
			}
			if got := sti.gtNonce(tt.nonce); got != tt.want {
				t.Errorf("StackTraceItem.gtNonce() = %v, want %v", got, tt.want)
			}
		})
	}
}
