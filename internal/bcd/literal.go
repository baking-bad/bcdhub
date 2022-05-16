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

// IsContractLazy -
func IsContractLazy(str string) bool {
	return len(str) == 36 && strings.HasPrefix(str, "KT1")
}

// IsAddressLazy -
func IsAddressLazy(address string) bool {
	return (len(address) == 36 && (strings.HasPrefix(address, "KT") || strings.HasPrefix(address, "tz"))) ||
		(len(address) == 37 && strings.HasPrefix(address, "txr"))
}

// IsRollupAddressLazy -
func IsRollupAddressLazy(address string) bool {
	return len(address) == 37 && strings.HasPrefix(address, "txr")
}

var (
	addressRegex   = regexp.MustCompile("(tz1|tz2|tz3|KT1|txr)[0-9A-Za-z]{33}")
	contractRegex  = regexp.MustCompile("(KT1)[0-9A-Za-z]{33}")
	bakerHashRegex = regexp.MustCompile("(SG1)[0-9A-Za-z]{33}")
)

// IsAddress -
func IsAddress(str string) bool {
	return addressRegex.MatchString(str)
}

// IsContract -
func IsContract(str string) bool {
	return contractRegex.MatchString(str)
}

// IsBakerHash -
func IsBakerHash(str string) bool {
	return bakerHashRegex.MatchString(str)
}
