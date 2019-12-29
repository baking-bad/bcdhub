package contractparser

import "regexp"

func findHardcodedAddresses(script string) []string {
	regexString := "(tz1|KT1)[0-9A-Za-z]{33}"
	re := regexp.MustCompile(regexString)
	return re.FindAllString(script, -1)
}
