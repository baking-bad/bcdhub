package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTags_Set(t *testing.T) {
	tests := []struct {
		name string
		t    Tags
		flag Tags
		want Tags
	}{
		{
			name: "test 1",
			t:    Tags(0),
			flag: ChainAwareTag | DelegatableTag,
			want: ChainAwareTag | DelegatableTag,
		}, {
			name: "test 2",
			t:    ChainAwareTag,
			flag: ChainAwareTag | DelegatableTag,
			want: ChainAwareTag | DelegatableTag,
		}, {
			name: "test 3",
			t:    ChainAwareTag | DelegatorTag,
			flag: ChainAwareTag | DelegatableTag,
			want: ChainAwareTag | DelegatableTag | DelegatorTag,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.t.Set(tt.flag)
			assert.Equal(t, tt.want, tt.t)
		})
	}
}

func TestTags_Clear(t *testing.T) {
	tests := []struct {
		name string
		t    Tags
		flag Tags
		want Tags
	}{
		{
			name: "test 1",
			t:    ChainAwareTag,
			flag: ChainAwareTag | DelegatableTag,
			want: 0,
		}, {
			name: "test 2",
			t:    DelegatableTag | UpgradableTag,
			flag: ChainAwareTag | DelegatableTag,
			want: UpgradableTag,
		}, {
			name: "test 3",
			t:    UpgradableTag,
			flag: ChainAwareTag | DelegatableTag,
			want: UpgradableTag,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.t.Clear(tt.flag)
			assert.Equal(t, tt.want, tt.t)
		})
	}
}

func TestTags_Has(t *testing.T) {
	tests := []struct {
		name string
		t    Tags
		flag Tags
		want bool
	}{
		{
			name: "test 1",
			t:    ChainAwareTag,
			flag: ChainAwareTag | DelegatableTag,
			want: true,
		}, {
			name: "test 2",
			t:    DelegatableTag | UpgradableTag,
			flag: ChainAwareTag | DelegatableTag,
			want: true,
		}, {
			name: "test 3",
			t:    UpgradableTag,
			flag: ChainAwareTag | DelegatableTag,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.Has(tt.flag); got != tt.want {
				t.Errorf("Tags.Has() = %v, want %v", got, tt.want)
			}
		})
	}
}
