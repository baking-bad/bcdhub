package contractparser

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/contractparser/language"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/tidwall/gjson"
)

func TestLanguage(t *testing.T) {
	testCases := []struct {
		name       string
		langsCode  helpers.Set
		langsParam helpers.Set
		expected   string
		err        error
	}{
		{
			name:       "Both unknown",
			langsCode:  helpers.Set{language.LangUnknown: struct{}{}},
			langsParam: helpers.Set{language.LangUnknown: struct{}{}},
			expected:   language.LangUnknown,
			err:        nil,
		},
		{
			name:       "Both detected",
			langsCode:  helpers.Set{language.LangLigo: struct{}{}},
			langsParam: helpers.Set{language.LangLigo: struct{}{}},
			expected:   language.LangLigo,
			err:        nil,
		},
		{
			name:       "Lang in Code",
			langsCode:  helpers.Set{language.LangLigo: struct{}{}},
			langsParam: helpers.Set{language.LangUnknown: struct{}{}},
			expected:   language.LangLigo,
			err:        nil,
		},
		{
			name:       "Lang in Parameter",
			langsCode:  helpers.Set{language.LangUnknown: struct{}{}},
			langsParam: helpers.Set{language.LangLigo: struct{}{}},
			expected:   language.LangLigo,
			err:        nil,
		},
		{
			name:       "Different lang in Code and Parameter",
			langsCode:  helpers.Set{language.LangLiquidity: struct{}{}},
			langsParam: helpers.Set{language.LangLigo: struct{}{}},
			expected:   language.LangLiquidity,
			err:        nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s := new(Script)
			s.Code.Code = gjson.Parse(`{}`)
			s.Code.Language = tt.langsCode
			s.Code.Parameter.Language = tt.langsParam

			lang, err := s.Language()
			if err != nil {
				t.Errorf("error Got: %s\n", err)
			}
			if lang != tt.expected {
				t.Errorf("wrong language detection\nGot: %s\nExpected: %s", lang, tt.expected)
			}
		})
	}
}
