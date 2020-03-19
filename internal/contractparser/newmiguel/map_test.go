package newmiguel

import "testing"

func Test_mapDecoder_getKey(t *testing.T) {
	tests := []struct {
		name    string
		key     *Node
		wantS   string
		wantErr bool
	}{
		{
			name:    "test int",
			key:     &Node{Value: 1},
			wantS:   "1",
			wantErr: false,
		}, {
			name:    "test int64",
			key:     &Node{Value: int64(64)},
			wantS:   "64",
			wantErr: false,
		}, {
			name:    "test string",
			key:     &Node{Value: "test string"},
			wantS:   "test string",
			wantErr: false,
		}, {
			name: "test array",
			key: &Node{
				Value: []interface{}{
					map[string]interface{}{"miguel_type": "string", "miguel_value": "hello"},
					map[string]interface{}{"miguel_type": "nat", "miguel_value": 42},
				}},
			wantS:   "hello@42",
			wantErr: false,
		}, {
			name:    "test map",
			key:     &Node{Value: map[string]interface{}{"miguel_type": "string", "miguel_value": "hello"}},
			wantS:   "hello",
			wantErr: false,
		}, {
			name:    "test error",
			key:     &Node{Value: 21.35},
			wantS:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &mapDecoder{}
			gotS, err := l.getKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("mapDecoder.getKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotS != tt.wantS {
				t.Errorf("mapDecoder.getKey() = %v, want %v", gotS, tt.wantS)
			}
		})
	}
}
