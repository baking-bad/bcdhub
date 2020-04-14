package contractparser

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/contractparser/language"

	"github.com/tidwall/gjson"
)

func TestDetectLorentCast(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "lorentz",
			input: `[[{"prim":"CAST"}]]`,
			want:  language.LangLorentz,
		},
		{
			name:  "michelson",
			input: `[[{"prim":"pair"}]]`,
			want:  language.LangUnknown,
		},
		{
			name:  "michelson",
			input: `[{"prim": "CAST"},{"prim": "bool"}]`,
			want:  language.LangUnknown,
		},
		{
			name:  "michelson",
			input: `[[{"prim": "nat"},{"prim": "CAST"}]]`,
			want:  language.LangUnknown,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			parsed := gjson.Parse(tt.input)
			if got := detectLorentCast(parsed); got != tt.want {
				t.Errorf("detectLorentCast invalid. expected: %v, got: %v", tt.want, got)
			}
		})
	}
}
