package contractparser

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"regexp"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/helpers"
)

// FindHardcodedAddresses -
func FindHardcodedAddresses(script fmt.Stringer) (helpers.Set, error) {
	s := script.String()
	regexString := "(tz|KT)[0-9A-Za-z]{34}"
	re := regexp.MustCompile(regexString)
	res := re.FindAllString(s, -1)
	resp := make(helpers.Set)
	resp.Append(res...)
	return resp, nil
}

// IsLiteral -
func IsLiteral(prim string) bool {
	return helpers.StringInArray(prim, []string{
		consts.CONTRACT, consts.BYTES, consts.ADDRESS, consts.KEYHASH,
		consts.KEY, consts.TIMESTAMP, consts.BOOL, consts.MUTEZ,
		consts.NAT, consts.STRING, consts.INT, consts.SIGNATURE,
	})
}

// ComputeContractHash -
func ComputeContractHash(code string) (string, error) {
	sha := sha512.New()
	if _, err := sha.Write([]byte(code)); err != nil {
		return "", err
	}
	return hex.EncodeToString(sha.Sum(nil)), nil
}
