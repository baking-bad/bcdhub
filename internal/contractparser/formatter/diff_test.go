package formatter

import (
	"reflect"
	"strings"
	"testing"
)

func Test_getLineSide(t *testing.T) {
	type args struct {
		res [][]Item
		i   int
	}
	tests := []struct {
		name  string
		args  args
		want  []Item
		want1 int
	}{
		{
			name: "Invalid index",
			args: args{
				res: make([][]Item, 0),
				i:   1,
			},
			want:  []Item{},
			want1: 0,
		}, {
			name: "Positive sum",
			args: args{
				res: [][]Item{
					[]Item{
						Item{
							Type: 1,
						}, Item{
							Type: 1,
						}, Item{
							Type: -1,
						},
					},
				},
				i: 0,
			},
			want: []Item{
				Item{
					Type: 1,
				}, Item{
					Type: 1,
				}, Item{
					Type: -1,
				},
			},
			want1: 1,
		}, {
			name: "Negative sum",
			args: args{
				res: [][]Item{
					[]Item{
						Item{
							Type: 1,
						}, Item{
							Type: -1,
						}, Item{
							Type: -1,
						},
					},
				},
				i: 0,
			},
			want: []Item{
				Item{
					Type: 1,
				}, Item{
					Type: -1,
				}, Item{
					Type: -1,
				},
			},
			want1: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getLineSide(tt.args.res, tt.args.i)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getLineSide() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getLineSide() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

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
