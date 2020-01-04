package contractparser

import "regexp"

import "encoding/json"

// FindHardcodedAddresses -
func FindHardcodedAddresses(script map[string]interface{}) ([]string, error) {
	b, err := json.Marshal(script)
	if err != nil {
		return nil, err
	}
	regexString := "(tz1|KT1)[0-9A-Za-z]{33}"
	re := regexp.MustCompile(regexString)
	return re.FindAllString(string(b), -1), nil
}
