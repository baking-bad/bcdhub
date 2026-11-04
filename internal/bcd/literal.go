package bcd

import (
	"regexp"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/encoding"
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
	return (len(address) == 36 && (strings.HasPrefix(address, "KT") ||
		strings.HasPrefix(address, "tz") ||
		strings.HasPrefix(address, "sr1"))) ||
		(len(address) == 37 && strings.HasPrefix(address, "txr"))
}

// IsRollupAddressLazy -
func IsRollupAddressLazy(address string) bool {
	return len(address) == 37 && strings.HasPrefix(address, "txr")
}

// IsRollupAddressLazy -
func IsSmartRollupAddressLazy(address string) bool {
	return len(address) == 36 && strings.HasPrefix(address, "sr1")
}

// base58 matches one character of the alphabet of Tezos base58check encodings.
const base58 = "[" + encoding.Base58Alphabet + "]"

// The validators below must match the whole input: unanchored, they accepted
// any string containing an address, and API parameters validated by them reach
// node RPC URL paths (e.g. `source` of run_operation).
var (
	addressRegex      = regexp.MustCompile("^(?:(?:tz[1-5]|KT1|sr1)" + base58 + "{33}|txr1" + base58 + "{33})$")
	contractRegex     = regexp.MustCompile("^KT1" + base58 + "{33}$")
	bakerHashRegex    = regexp.MustCompile("^SG1" + base58 + "{33}$")
	operationRegex    = regexp.MustCompile("^o" + base58 + "{50}$")
	smartRollupRegex  = regexp.MustCompile("^sr1" + base58 + "{33}$")
	findContractRegex = regexp.MustCompile("KT1" + base58 + "{33}")
)

// IsAddress -
func IsAddress(str string) bool {
	return addressRegex.MatchString(str)
}

// IsContract -
func IsContract(str string) bool {
	return contractRegex.MatchString(str)
}

// FindContract -
func FindContract(str string) string {
	return findContractRegex.FindString(str)
}

// IsBakerHash -
func IsBakerHash(str string) bool {
	return bakerHashRegex.MatchString(str)
}

// IsOperationHash -
func IsOperationHash(str string) bool {
	return operationRegex.MatchString(str)
}

// IsSmartRollupHash -
func IsSmartRollupHash(str string) bool {
	return smartRollupRegex.MatchString(str)
}
