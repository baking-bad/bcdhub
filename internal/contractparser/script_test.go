package contractparser

import (
	"fmt"
	"testing"

	"github.com/baking-bad/bcdhub/internal/contractparser/language"
)

func TestLanguage(t *testing.T) {
	testCases := []struct {
		name      string
		langCode  string
		langParam string
		expected  string
		err       error
	}{
		{
			name:      "Both unknown",
			langCode:  language.LangUnknown,
			langParam: language.LangUnknown,
			expected:  language.LangUnknown,
			err:       nil,
		},
		{
			name:      "Both detected",
			langCode:  language.LangLigo,
			langParam: language.LangLigo,
			expected:  language.LangLigo,
			err:       nil,
		},
		{
			name:      "Lang in Code",
			langCode:  language.LangLigo,
			langParam: language.LangUnknown,
			expected:  language.LangLigo,
			err:       nil,
		},
		{
			name:      "Lang in Parameter",
			langCode:  language.LangUnknown,
			langParam: language.LangLigo,
			expected:  language.LangLigo,
			err:       nil,
		},
		{
			name:      "Different lang in Code and Parameter",
			langCode:  language.LangLiquidity,
			langParam: language.LangLigo,
			expected:  "",
			err:       fmt.Errorf("Language detect error. [code] %s | [parameter] %s", language.LangLiquidity, language.LangLigo),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s := new(Script)
			s.Code.Language = tt.langCode
			s.Code.Parameter.Language = tt.langParam

			lang, err := s.Language()
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("wrong error\nGot: %s\nExpected: %s", err, tt.err)
			}
			if lang != tt.expected {
				t.Errorf("wrong language detection\nGot: %s\nExpected: %s", lang, tt.expected)
			}
		})
	}
}
