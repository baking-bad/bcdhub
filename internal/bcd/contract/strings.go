package contract

import (
	"crypto/sha512"
	"encoding/hex"
	"regexp"

	"github.com/baking-bad/bcdhub/internal/bcd/types"
)

func computeHash(data []byte) (string, error) {
	sha := sha512.New()
	if _, err := sha.Write(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(sha.Sum(nil)), nil
}

func findHardcodedAddresses(code []byte) (types.Set, error) {
	regexString := "(tz|KT)[0-9A-Za-z]{34}"
	re := regexp.MustCompile(regexString)
	res := re.FindAllString(string(code), -1)
	resp := make(types.Set)
	resp.Append(res...)
	return resp, nil
}

// IsAddress -
func IsAddress(str string) bool {
	regexString := "(tz|KT)[0-9A-Za-z]{34}"
	re := regexp.MustCompile(regexString)
	return re.MatchString(str)
}
