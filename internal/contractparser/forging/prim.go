package forging

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
	"__CREATE_ACCOUNT__", // DEPRECATED
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
	"DIG",
	"DUG",
	"EMPTY_BIG_MAP",
	"APPLY",
	"chain_id",
	"CHAIN_ID",
	// EDO
	"LEVEL",
	"SELF_ADDRESS",
	"never",
	"NEVER",
	"UNPAIR",
	"VOTING_POWER",
	"TOTAL_VOTING_POWER",
	"KECCAK",
	"SHA3",
	"PAIRING_CHECK",
	"bls12_381_g1",
	"bls12_381_g2",
	"bls12_381_fr",
	"sapling_state",
	"sapling_transaction",
	"SAPLING_EMPTY_STATE",
	"SAPLING_VERIFY_UPDATE",
	"ticket",
	"TICKET",
	"READ_TICKET",
	"SPLIT_TICKET",
	"JOIN_TICKETS",
	"GET_AND_UPDATE",
}

type primForger struct {
	ArgsCount int
	HasAnnots bool
}

func newPrimForger(argsCount int, hasAnnots bool) primForger {
	return primForger{
		ArgsCount: argsCount,
		HasAnnots: hasAnnots,
	}
}

// Decode -
func (d primForger) Unforge(dec *decoder, code *strings.Builder) (length int, err error) {
	prim, err := unforgePrim(dec)
	if err != nil {
		return 1, err
	}
	fmt.Fprintf(code, `{ "prim": "%s"`, prim)
	length++

	// log.Printf("[primDecoder Decode] data: %d | args count: %d | has annots: %v", length, d.ArgsCount, d.HasAnnots)
	if d.ArgsCount > 0 {
		fmt.Fprintf(code, `, "args": [ `)
		n, err := unforgeArgs(dec, code, d.ArgsCount)
		if err != nil {
			return n + 1, err
		}
		fmt.Fprintf(code, ` ]`)
		length += n
	}

	if d.HasAnnots {
		fmt.Fprintf(code, `, "annots": [ "`)

		annots, n, err := unforgeAnnots(dec)
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
func (d primGeneral) Unforge(dec *decoder, code *strings.Builder) (length int, err error) {
	prim, err := unforgePrim(dec)
	if err != nil {
		return 1, err
	}
	fmt.Fprintf(code, `{ "prim": "%s", "args": `, prim)
	length++

	ad := arrayForger{}
	n, err := ad.Unforge(dec, code)
	if err != nil {
		return n + 1, err
	}
	length += n

	annots, n, err := unforgeAnnots(dec)
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
