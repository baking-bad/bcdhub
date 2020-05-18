package migrations

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/language"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
)

// SetLanguage - migration that set langugage to contract by annotations or entrypoints
type SetLanguage struct{}

// Key -
func (m *SetLanguage) Key() string {
	return "language"
}

// Description -
func (m *SetLanguage) Description() string {
	return "set langugage to contract by annotations or entrypoints"
}

// Do - migrate function
func (m *SetLanguage) Do(ctx *config.Context) error {
	filter := make(map[string]interface{})
	filter["language"] = language.LangUnknown

	contracts, err := ctx.ES.GetContracts(filter)
	if err != nil {
		return err
	}

	logger.Info("Found %d contracts", len(contracts))

	for i, c := range contracts {
		lang := getLanguage(c.FailStrings, c.Annotations, c.Entrypoints)

		if lang == language.LangUnknown {
			continue
		}

		c.Language = lang

		if _, err := ctx.ES.UpdateDoc(elastic.DocContracts, c.ID, c); err != nil {
			log.Println("ctx.ES.UpdateDoc error:", c.ID, c, err)
			return err
		}

		log.Printf("%d/%d | %v | [%v]", i, len(contracts), c.ID, lang)
	}

	return nil
}

type liquidity struct{}
type ligo struct{}
type lorentz struct{}
type smartpy struct{}
type detector interface {
	Detect([]string, []string, []string) bool
}

func getLanguage(failstrings, annotations, entrypoints []string) string {
	languages := map[string]detector{
		language.LangSmartPy:   smartpy{},
		language.LangLiquidity: liquidity{},
		language.LangLigo:      ligo{},
		language.LangLorentz:   lorentz{},
	}

	for language, l := range languages {
		if l.Detect(failstrings, annotations, entrypoints) {
			return language
		}
	}

	return language.LangUnknown
}

func (l liquidity) Detect(_, annotations, entrypoints []string) bool {
	for _, a := range annotations {
		if strings.Contains(a, "_slash_") || strings.Contains(a, ":_entries") || strings.Contains(a, `@\w+_slash_1`) {
			return true
		}
	}

	for _, e := range entrypoints {
		if strings.Contains(e, "_Liq_entry") {
			return true
		}
	}

	return false
}

func (l ligo) Detect(failstrings, annotations, _ []string) bool {
	for _, a := range annotations {
		if len(a) < 2 {
			continue
		}
		if a[0] == '%' && isDigit(a[1:]) {
			return true
		}
	}

	for _, f := range failstrings {
		if hasLIGOKeywords(f) {
			return true
		}
	}

	return false
}

func isDigit(input string) bool {
	_, err := strconv.ParseUint(input, 10, 32)
	return err == nil
}

func hasLIGOKeywords(s string) bool {
	ligoKeywords := []string{
		"GET_FORCE",
		"get_force",
		"MAP FIND",
	}

	for _, keyword := range ligoKeywords {
		if s == keyword {
			return true
		}
	}

	return strings.Contains(s, "get_entrypoint") || strings.Contains(s, "get_contract")
}

var lorentzCamelCase = regexp.MustCompile(`([A-Z][a-z0-9]+)((\d)|([A-Z0-9][a-z0-9]+))*([A-Z])?`)

func (l lorentz) Detect(failstrings, _, entrypoints []string) bool {
	for _, f := range failstrings {
		if strings.Contains(f, "UStore") {
			return true
		}
	}

	for _, e := range entrypoints {
		if strings.HasPrefix(e, "epw") && lorentzCamelCase.MatchString(e[3:]) {
			return true
		}
	}

	return false
}

func (l smartpy) Detect(failstrings, _, _ []string) bool {
	for _, f := range failstrings {
		if strings.Contains(f, "SmartPy") ||
			strings.Contains(f, "self.") ||
			strings.Contains(f, "sp.") ||
			strings.Contains(f, "WrongCondition") ||
			strings.Contains(f, `Get-item:\d+`) {
			return true
		}
	}

	return false
}
