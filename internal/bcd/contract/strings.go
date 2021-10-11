package contract

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
)

var (
	regAddress = regexp.MustCompile("(tz|KT)[0-9A-Za-z]{34}")
)

// ComputeHash -
func ComputeHash(data []byte) (string, error) {
	sha := sha256.New()
	if _, err := sha.Write(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(sha.Sum(nil)), nil
}

// IsAddress -
func IsAddress(str string) bool {
	return len(str) == 36 && regAddress.MatchString(str)
}
