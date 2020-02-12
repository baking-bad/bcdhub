package metrics

import "testing"

func Test_distance(t *testing.T) {
	type args struct {
		a          string
		b          string
		threashold uint16
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "simple 1",
			args: args{
				a:          "1",
				b:          "11",
				threashold: 2,
			},
			want: 1,
		}, {
			name: "simple 2",
			args: args{
				a:          "11",
				b:          "11",
				threashold: 2,
			},
			want: 0,
		}, {
			name: "simple 3",
			args: args{
				a:          "",
				b:          "11",
				threashold: 2,
			},
			want: 2,
		}, {
			name: "simple 4",
			args: args{
				a:          "1234",
				b:          "5678",
				threashold: 2,
			},
			want: 3,
		}, {
			name: "simple 5",
			args: args{
				a:          "12345",
				b:          "12",
				threashold: 1,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := distance(tt.args.a, tt.args.b, tt.args.threashold); got != tt.want {
				t.Errorf("distance() = %v, want %v", got, tt.want)
			}
		})
	}
}
