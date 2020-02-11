package rawbytes

import (
	"fmt"
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

type primDecoder struct {
	ArgsCount int
	HasAnnots bool
}

func newPrimDecoder(argsCount int, hasAnnots bool) primDecoder {
	return primDecoder{
		ArgsCount: argsCount,
		HasAnnots: hasAnnots,
	}
}

// Decode -
func (d primDecoder) Decode(dec *decoder, code *strings.Builder) (length int, err error) {
	prim, err := decodePrim(dec)
	if err != nil {
		return 1, err
	}
	fmt.Fprintf(code, `{ "prim": "%s"`, prim)
	length++

	if d.ArgsCount > 0 {
		fmt.Fprintf(code, `, "args": [ `)
		n, err := decodeArgs(dec, code, d.ArgsCount)
		if err != nil {
			return n + 1, err
		}
		fmt.Fprintf(code, ` ]`)
		length += n
	}

	if d.HasAnnots {
		fmt.Fprintf(code, `, "annots": [ "`)

		annots, n, err := decodeAnnots(dec)
		if err != nil {
			return length + 1, err
		}
		fmt.Fprintf(code, "%s\" ]", annots)
		length += n
	}
	fmt.Fprintf(code, ` }`)
	return length, nil
}

type primGeneral struct{}

// Decode -
func (d primGeneral) Decode(dec *decoder, code *strings.Builder) (length int, err error) {
	prim, err := decodePrim(dec)
	if err != nil {
		return 1, err
	}
	fmt.Fprintf(code, `{ "prim": "%s", "args": `, prim)
	length++

	ad := arrayDecoder{}
	n, err := ad.Decode(dec, code)
	if err != nil {
		return n + 1, err
	}
	length += n

	annots, n, err := decodeAnnots(dec)
	if err != nil {
		return length + 1, err
	}

	length += n
	if n != 4 {
		fmt.Fprintf(code, `, "annots": [ "%s" ]`, annots)
	}

	fmt.Fprintf(code, ` }`)
	return length, nil
}
