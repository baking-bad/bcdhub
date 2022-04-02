package bcd

import (
	"regexp"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
)

// IsLiteral -
func IsLiteral(prim string) bool {
	for _, s := range []string{
		consts.CONTRACT, consts.BYTES, consts.ADDRESS, consts.KEYHASH,
		consts.KEY, consts.TIMESTAMP, consts.BOOL, consts.MUTEZ,
		consts.NAT, consts.STRING, consts.INT, consts.SIGNATURE,
	} {
		if prim == s {
			return true
		}
	}
	return false
}

// IsContract -
func IsContract(address string) bool {
	return len(address) == 36 && strings.HasPrefix(address, "KT")
}

// IsAddressLazy -
func IsAddressLazy(address string) bool {
	return len(address) == 36 && (strings.HasPrefix(address, "KT") || strings.HasPrefix(address, "tz"))
}

// IsAddress -
func IsAddress(str string) bool {
	regexString := "(tz|KT)[0-9A-Za-z]{34}"
	re := regexp.MustCompile(regexString)
	return re.MatchString(str)
}
