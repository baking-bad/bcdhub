package rawbytes

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

var primKeywords = []string{
	"parameter",
	"storage",
	"code",
	"False",
	"Elt",
	"Left",
	"None",
	"Pair",
	"Right",
	"Some",
	"True",
	"Unit",
	"PACK",
	"UNPACK",
	"BLAKE2B",
	"SHA256",
	"SHA512",
	"ABS",
	"ADD",
	"AMOUNT",
	"AND",
	"BALANCE",
	"CAR",
	"CDR",
	"CHECK_SIGNATURE",
	"COMPARE",
	"CONCAT",
	"CONS",
	"CREATE_ACCOUNT",
	"CREATE_CONTRACT",
	"IMPLICIT_ACCOUNT",
	"DIP",
	"DROP",
	"DUP",
	"EDIV",
	"EMPTY_MAP",
	"EMPTY_SET",
	"EQ",
	"EXEC",
	"FAILWITH",
	"GE",
	"GET",
	"GT",
	"HASH_KEY",
	"IF",
	"IF_CONS",
	"IF_LEFT",
	"IF_NONE",
	"INT",
	"LAMBDA",
	"LE",
	"LEFT",
	"LOOP",
	"LSL",
	"LSR",
	"LT",
	"MAP",
	"MEM",
	"MUL",
	"NEG",
	"NEQ",
	"NIL",
	"NONE",
	"NOT",
	"NOW",
	"OR",
	"PAIR",
	"PUSH",
	"RIGHT",
	"SIZE",
	"SOME",
	"SOURCE",
	"SENDER",
	"SELF",
	"STEPS_TO_QUOTA",
	"SUB",
	"SWAP",
	"TRANSFER_TOKENS",
	"SET_DELEGATE",
	"UNIT",
	"UPDATE",
	"XOR",
	"ITER",
	"LOOP_LEFT",
	"ADDRESS",
	"CONTRACT",
	"ISNAT",
	"CAST",
	"RENAME",
	"bool",
	"contract",
	"int",
	"key",
	"key_hash",
	"lambda",
	"list",
	"map",
	"big_map",
	"nat",
	"option",
	"or",
	"pair",
	"set",
	"signature",
	"string",
	"bytes",
	"mutez",
	"timestamp",
	"unit",
	"operation",
	"address",
	"SLICE",
	"DEFAULT_ACCOUNT",
	"tez",
}

type primDecoder struct{}

// Decode -
func (d primDecoder) Decode(dec io.Reader, code *strings.Builder) (int, error) {
	b := make([]byte, 4)
	if n, err := dec.Read(b); err != nil {
		return n, err
	}
	key := int(binary.LittleEndian.Uint32(b))
	if key > len(primKeywords) {
		return 4, fmt.Errorf("invalid prim keyword %s", b)
	}
	fmt.Fprintf(code, `{ "prim": "%s" }`, primKeywords[key])
	return 4, nil
}

type primAnnotsDecoder struct{}

// Decode -
func (d primAnnotsDecoder) Decode(dec io.Reader, code *strings.Builder) (int, error) {
	b := make([]byte, 1)
	if n, err := dec.Read(b); err != nil {
		return n, err
	}
	key := int(b[0])
	if key > len(primKeywords) {
		return 1, fmt.Errorf("invalid prim keyword %s", b)
	}
	fmt.Fprintf(code, `{ "prim": "%s", "annots": [ "`, primKeywords[key])

	sb := make([]byte, 4)
	if n, err := dec.Read(sb); err != nil {
		return n, err
	}

	length := int(binary.BigEndian.Uint32(sb))
	data := make([]byte, length)
	if _, err := dec.Read(data); err != nil && err != io.EOF {
		return 1 + 4 + length, err
	}

	var ret []string
	for _, v := range strings.Split(string(data), " ") {
		ret = append(ret, v)
	}

	annots := strings.Join(ret, "\",\"")
	fmt.Fprintf(code, "%s\" ] }", annots)
	return length + 4 + 1, nil
}

type primArgsDecoder struct{}

// Decode -
func (d primArgsDecoder) Decode(dec io.Reader, code *strings.Builder) (int, error) {
	b := make([]byte, 1)
	if n, err := dec.Read(b); err != nil {
		return n, err
	}
	key := int(b[0])
	if key > len(primKeywords) {
		return 1, fmt.Errorf("invalid prim keyword %s", b)
	}
	fmt.Fprintf(code, `{ "prim": "%s", "args": [ `, primKeywords[key])

	n, err := hexToMicheline(dec, code)
	if err != nil {
		return n + 1, err
	}

	fmt.Fprintf(code, ` ] }`)
	return n + 1, nil
}
