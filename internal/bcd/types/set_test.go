package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet_Add(t *testing.T) {
	tests := []struct {
		name    string
		items   []string
		wantLen int
	}{
		{
			name:    "Set 1 item",
			items:   []string{"test"},
			wantLen: 1,
		}, {
			name:    "Set 2 items",
			items:   []string{"test", "value"},
			wantLen: 2,
		}, {
			name:    "Set 2 similar items",
			items:   []string{"test", "test"},
			wantLen: 1,
		},
	}
	for _, tt := range tests {
		s := make(Set)
		t.Run(tt.name, func(t *testing.T) {
			for i := range tt.items {
				s.Add(tt.items[i])
			}
			assert.Equal(t, tt.wantLen, s.Len())
		})
	}
}

func TestSet_Append(t *testing.T) {
	tests := []struct {
		name    string
		items   []string
		wantLen int
	}{
		{
			name:    "Set 1 item",
			items:   []string{"test"},
			wantLen: 1,
		}, {
			name:    "Set 2 items",
			items:   []string{"test", "value"},
			wantLen: 2,
		}, {
			name:    "Set 2 similar items",
			items:   []string{"test", "test", ""},
			wantLen: 1,
		},
	}
	for _, tt := range tests {
		s := make(Set)
		t.Run(tt.name, func(t *testing.T) {
			s.Append(tt.items...)
		})
		assert.Equal(t, tt.wantLen, s.Len())
	}
}

func TestSet_Values(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			name: "items",
			want: []string{"test", "test2", "test3"},
		},
	}
	for _, tt := range tests {
		s := make(Set)
		t.Run(tt.name, func(t *testing.T) {
			s.Append(tt.want...)
			assert.ElementsMatch(t, s.Values(), tt.want)
		})
	}
}

func TestSet_Merge(t *testing.T) {
	tests := []struct {
		name string
		one  []string
		two  []string
		want []string
	}{
		{
			name: "test",
			one:  []string{"test", "", "test2", "test3"},
			two:  []string{"test2", "", "test3", "test4"},
			want: []string{"test", "test2", "test3", "test4"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			one := make(Set)
			one.Append(tt.one...)
			two := make(Set)
			two.Append(tt.two...)
			one.Merge(two)

			assert.ElementsMatch(t, one.Values(), tt.want)
		})
	}
}
