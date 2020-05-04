package formatter

import (
	"strings"
	"testing"
)

func Test_skipSpaces(t *testing.T) {
	type args struct {
		s      string
		offset int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Without spaces",
			args: args{
				s:      "123456",
				offset: 0,
			},
			want: 0,
		}, {
			name: "2 spaces",
			args: args{
				s:      "  123456",
				offset: 0,
			},
			want: 2,
		}, {
			name: "2 spaces with 1 offset",
			args: args{
				s:      "  123456",
				offset: 1,
			},
			want: 2,
		}, {
			name: "Only spaces",
			args: args{
				s:      strings.Repeat(" ", 10),
				offset: 3,
			},
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := skipSpaces(tt.args.s, tt.args.offset); got != tt.want {
				t.Errorf("skipSpaces() = %v, want %v", got, tt.want)
			}
		})
	}
}
