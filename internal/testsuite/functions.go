package testsuite

import "encoding/hex"

func Ptr[T any](val T) *T {
	return &val
}

func MustHexDecode(s string) []byte {
	data, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}
