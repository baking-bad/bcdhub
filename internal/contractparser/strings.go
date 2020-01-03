package contractparser

import "regexp"

// FindHardcodedAddresses -
func FindHardcodedAddresses(script string) []string {
	regexString := "(tz1|KT1)[0-9A-Za-z]{33}"
	re := regexp.MustCompile(regexString)
	return re.FindAllString(script, -1)
}
