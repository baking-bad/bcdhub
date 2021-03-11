package contract

import (
	"crypto/sha512"
	"encoding/hex"
	"regexp"

	"github.com/baking-bad/bcdhub/internal/bcd/types"
)

var (
	regAddress = regexp.MustCompile("(tz|KT)[0-9A-Za-z]{34}")
)

// ComputeHash -
func ComputeHash(data []byte) (string, error) {
	sha := sha512.New()
	if _, err := sha.Write(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(sha.Sum(nil)), nil
}

func findHardcodedAddresses(code []byte) types.Set {
	res := regAddress.FindAllString(string(code), -1)
	resp := make(types.Set)
	resp.Append(res...)
	return resp
}

// IsAddress -
func IsAddress(str string) bool {
	return len(str) == 36 && regAddress.MatchString(str)
}
