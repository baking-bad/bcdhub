package stringer

import (
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"

	jsoniter "github.com/json-iterator/go"

	"github.com/baking-bad/bcdhub/internal/contractparser/forging"
	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/baking-bad/bcdhub/internal/contractparser/unpack"
	"github.com/tidwall/gjson"
)

const minPrintableASCII = 32

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type arg struct {
	Value interface{} `json:"-"`
}

type node struct {
	Args       []arg    `json:"args,omitempty"`
	Prim       string   `json:"prim,omitempty"`
	Annots     []string `json:"annots,omitempty"`
	String     string   `json:"string,omitempty"`
	Bytes      string   `json:"bytes,omitempty"`
	Entrypoint string   `json:"entrypoint,omitempty"`
	Value      *node    `json:"value,omitempty"`
}

// UnmarshalJSON -
func (a *arg) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	switch data[0] {
	case '[':
		var arr []node
		if err := json.Unmarshal(data, &arr); err != nil {
			return err
		}
		a.Value = arr
	case '{':
		var obj node
		if err := json.Unmarshal(data, &obj); err != nil {
			return err
		}
		a.Value = obj
	}
	return nil
}

// Get - returnes slice of unique meaningful strings from json
func Get(data string) []string {
	var storage = make(map[string]struct{})

	if err := parse(data, storage); err != nil {
		return nil
	}

	var result = make([]string, 0, len(storage))
	for key := range storage {
		result = append(result, key)
	}

	return result
}

func parse(data string, storage map[string]struct{}) error {
	if strings.HasPrefix(data, "{") {
		var obj node
		if err := json.Unmarshal([]byte(data), &obj); err != nil {
			return err
		}
		findInObject(obj, storage)
	} else if strings.HasPrefix(data, "[") {
		var arr []node
		if err := json.Unmarshal([]byte(data), &arr); err != nil {
			return err
		}
		findInArray(arr, storage)
	}
	return nil
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

func findStrings(obj interface{}, storage map[string]struct{}) {
	switch val := obj.(type) {
	case node:
		findInObject(val, storage)
	case []node:
		findInArray(val, storage)
	}
}

func findInArray(val []node, storage map[string]struct{}) {
	for _, n := range val {
		findStrings(n, storage)
	}
}

func findInObject(val node, storage map[string]struct{}) {
	if val.String != "" {
		findInString(val.String, storage)
		return
	}

	if val.Bytes != "" {
		findInBytes(val.Bytes, storage)
		return
	}

	for _, arg := range val.Args {
		findStrings(arg.Value, storage)
	}

	if val.Entrypoint != "" && val.Value != nil {
		findStrings(val.Value, storage)
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
		str, err := forging.Unforge(input[2:])
		if err == nil {
			if err := parse(str, storage); err != nil {
				return
			}
		}
	}

	decoded, err := hex.DecodeString(input)
	if err == nil && isASCII(decoded) {
		storage[string(decoded)] = struct{}{}
	}
}

func isASCII(input []byte) bool {
	for _, c := range input {
		if c < minPrintableASCII || c > unicode.MaxASCII {
			return false
		}
	}
	return true
}
