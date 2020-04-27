package helpers

import "regexp"

// IsAddress -
func IsAddress(str string) bool {
	regexString := "(tz|KT)[0-9A-Za-z]{34}"
	re := regexp.MustCompile(regexString)
	return re.MatchString(str)
}
