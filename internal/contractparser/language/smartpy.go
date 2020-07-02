package language

import (
	"regexp"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/node"
	"github.com/tidwall/gjson"
)

type smartpy struct{}

func (l smartpy) DetectInCode(n node.Node) bool {
	str := n.GetString()

	if str == "" {
		return false
	}

	if strings.HasPrefix(str, "Get-item:") {
		re := regexp.MustCompile(`^Get-item:\d+$`)
		return re.MatchString(str)
	}

	return strings.Contains(str, "SmartPy") ||
		strings.Contains(str, "self.") ||
		strings.Contains(str, "sp.") ||
		strings.Contains(str, "WrongCondition")
}

func (l smartpy) DetectInParameter(n node.Node) bool {
	return false
}

func (l smartpy) DetectInFirstPrim(val gjson.Result) bool {
	return false
}
