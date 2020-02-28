package contractparser

import (
	"regexp"
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/helpers"
	"github.com/tidwall/gjson"
)

// FindHardcodedAddresses -
func FindHardcodedAddresses(script gjson.Result) (helpers.Set, error) {
	s := script.String()
	regexString := "(tz|KT)[0-9A-Za-z]{34}"
	re := regexp.MustCompile(regexString)
	res := re.FindAllString(s, -1)
	resp := make(helpers.Set)
	resp.Append(res...)
	return resp, nil
}

// IsLiteral -
func IsLiteral(prim string) bool {
	return helpers.StringInArray(prim, []string{
		consts.CONTRACT, consts.BYTES, consts.ADDRESS, consts.KEYHASH,
		consts.KEY, consts.TIMESTAMP, consts.BOOL, consts.MUTEZ,
		consts.NAT, consts.STRING, consts.INT, consts.SIGNATURE,
	})
}

// IsParametersError -
func IsParametersError(errorString string) bool {
	data := gjson.Parse(errorString)
	if !data.IsArray() {
		return false
	}
	for _, err := range data.Array() {
		errID := err.Get("id").String()
		if strings.Contains(errID, consts.BadParameterError) {
			return true
		}
	}
	return false
}
