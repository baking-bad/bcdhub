package language

import (
	"regexp"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
)

type smartpy struct{}

func (l smartpy) DetectInCode(node *base.Node) bool {
	if node.StringValue == nil {
		return false
	}

	str := *node.StringValue
	if strings.HasPrefix(str, "Get-item:") {
		re := regexp.MustCompile(`^Get-item:\d+$`)
		return re.MatchString(str)
	}

	return strings.Contains(str, "SmartPy") ||
		strings.Contains(str, "self.") ||
		strings.Contains(str, "sp.") ||
		strings.Contains(str, "WrongCondition")
}

func (l smartpy) DetectInParameter(node *base.Node) bool {
	return false
}

func (l smartpy) DetectInFirstPrim(node *base.Node) bool {
	return false
}
