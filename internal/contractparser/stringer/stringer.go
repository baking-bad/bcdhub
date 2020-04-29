package stringer

import (
	"encoding/hex"
	"encoding/json"
	"regexp"

	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/baking-bad/bcdhub/internal/contractparser/unpack"
	"github.com/baking-bad/bcdhub/internal/contractparser/unpack/rawbytes"
	"github.com/tidwall/gjson"
)

// Get - returnes slice of unique meaningful strings from json
func Get(node gjson.Result) []string {
	var storage = make(map[string]struct{})
	findStrings(node, storage)

	var result = make([]string, 0, len(storage))
	for key := range storage {
		result = append(result, key)
	}

	return result
}

// Stringify -
func Stringify(node gjson.Result) string {
	if node.IsObject() {
		if node.Get("string").Exists() {
			return node.Get("string").String()
		}

		if node.Get("bytes").Exists() {
			hex := node.Get("bytes").String()
			return unpackBytes(hex)
		}
	}

	if res, err := formatter.MichelineToMichelson(node, true, formatter.DefLineSize); err == nil {
		return res
	}

	return node.String()
}

// StringifyInterface -
func StringifyInterface(value interface{}) (string, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	g := gjson.ParseBytes(b)
	return Stringify(g), nil
}

func unpackBytes(bytes string) string {
	if res, err := unpack.KeyHash(bytes); err == nil {
		return res
	}

	if res, err := unpack.Contract(bytes); err == nil {
		return res
	}

	return unpack.Bytes(bytes)
}

func findStrings(node gjson.Result, storage map[string]struct{}) {
	if node.IsArray() {
		findInArray(node, storage)
	}

	if node.IsObject() {
		findInObject(node, storage)
	}
}

func findInArray(node gjson.Result, storage map[string]struct{}) {
	for _, n := range node.Array() {
		findStrings(n, storage)
	}
}

func findInObject(node gjson.Result, storage map[string]struct{}) {
	if node.Get("string").Exists() {
		findInString(node.Get("string").String(), storage)
		return
	}

	if node.Get("bytes").Exists() {
		findInBytes(node.Get("bytes").String(), storage)
		return
	}

	if node.Get("args").Exists() {
		for _, args := range node.Get("args").Array() {
			findStrings(args, storage)
		}
	}

	if node.Get("entrypoint").Exists() && node.Get("value").Exists() {
		value := node.Get("value")
		findStrings(value, storage)
	}
}

var regexpRFC3339 = regexp.MustCompile(`^([0-9]+)-(0[1-9]|1[012])-(0[1-9]|[12][0-9]|3[01])[Tt]([01][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9]|60)(\.[0-9]+)?(([Zz])|([\+|\-]([01][0-9]|2[0-3]):[0-5][0-9]))$`)

func findInString(input string, storage map[string]struct{}) {
	if !regexpRFC3339.MatchString(input) {
		storage[input] = struct{}{}
	}
}

func findInBytes(input string, storage map[string]struct{}) {
	if res, err := unpack.KeyHash(input); err == nil {
		storage[res] = struct{}{}
		return
	}

	if res, err := unpack.Address(input); err == nil {
		storage[res] = struct{}{}
		return
	}

	if res, err := unpack.Contract(input); err == nil {
		storage[res] = struct{}{}
		return
	}

	if len(input) >= 1 && input[:2] == unpack.MainPrefix {
		str, err := rawbytes.ToMicheline(input[2:])
		if err == nil {
			data := gjson.Parse(str)
			findStrings(data, storage)
		}
	}

	decoded, err := hex.DecodeString(input)
	if err == nil && unpack.IsASCII(decoded) {
		storage[string(decoded)] = struct{}{}
	}
}
