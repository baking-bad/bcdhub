package contractparser

import (
	"regexp"

	"github.com/tidwall/gjson"
)

// FindHardcodedAddresses -
func FindHardcodedAddresses(script gjson.Result) (Set, error) {
	s := script.String()
	regexString := "(tz|KT)[0-9A-Za-z]{34}"
	re := regexp.MustCompile(regexString)
	res := re.FindAllString(s, -1)
	resp := make(Set)
	resp.Append(res...)
	return resp, nil
}
