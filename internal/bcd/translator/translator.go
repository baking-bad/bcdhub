package translator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/yhirose/go-peg"
)

// MichelineTranslator -
type MichelineTranslator struct {
	handlers map[string]func(ast *peg.Ast) (string, error)
}

// NewJSONTranslator -
func NewJSONTranslator() *MichelineTranslator {
	t := MichelineTranslator{}
	t.handlers = map[string]func(ast *peg.Ast) (string, error){
		"instrs":        t.arrayTranslate,
		"instr":         t.pass,
		"expr":          t.exprTranslate,
		"prim":          t.tokenTranslate,
		"args":          t.arrayTranslate,
		"arg":           t.argTranslate,
		"Int":           t.intTranslate,
		"String":        t.stringTranslate,
		"StringContent": t.pass,
		"annots":        t.arrayTranslate,
		"annot":         t.tokenTranslate,
		"Byte":          t.bytesTranslate,
		"complex_instr": t.complexInstrTranslate,
	}
	return &t
}

// Translate -
func (t *MichelineTranslator) Translate(ast *peg.Ast) (string, error) {
	handler, ok := t.handlers[ast.Name]
	if ok {
		return handler(ast)
	}
	return t.pass(ast)
}

func (t *MichelineTranslator) pass(ast *peg.Ast) (string, error) {
	if len(ast.Nodes) > 0 {
		return t.Translate(ast.Nodes[0])
	}
	return "", nil
}

func (t *MichelineTranslator) exprTranslate(ast *peg.Ast) (string, error) {
	var s strings.Builder
	s.WriteByte('{')
	for i := range ast.Nodes {
		data, err := t.Translate(ast.Nodes[i])
		if err != nil {
			return "", err
		}
		if data != "" {
			if s.Len() > 1 {
				s.WriteByte(',')
			}
			if strings.HasPrefix(data, "[") || strings.HasPrefix(data, "{") {
				s.WriteString(fmt.Sprintf(`"%s":%s`, ast.Nodes[i].Name, data))
			} else {
				s.WriteString(fmt.Sprintf(`"%s":"%s"`, ast.Nodes[i].Name, data))
			}
		}
	}
	s.WriteByte('}')
	return s.String(), nil
}

func (t *MichelineTranslator) tokenTranslate(ast *peg.Ast) (string, error) {
	if ast.Name == "prim" {
		if err := validatePrimitive(ast.Token); err != nil {
			return "", err
		}
	}
	return ast.Token, nil
}

func (t *MichelineTranslator) arrayTranslate(ast *peg.Ast) (string, error) {
	var s strings.Builder
	s.WriteByte('[')

	var count int
	for i := range ast.Nodes {
		arg, err := t.Translate(ast.Nodes[i])
		if err != nil {
			return "", err
		}
		if arg != "" {
			if s.Len() > 1 {
				s.WriteByte(',')
			}
			if ast.Nodes[i].Name == "annot" {
				s.WriteString(fmt.Sprintf(`"%s"`, arg))
			} else {
				s.WriteString(arg)
			}
			count++
		}
	}
	s.WriteByte(']')
	return s.String(), nil
}

func (t *MichelineTranslator) argTranslate(ast *peg.Ast) (string, error) {
	for i := range ast.Nodes {
		if ast.Nodes[i].Name == "prim" {
			prim, err := t.Translate(ast.Nodes[i])
			if err != nil {
				return "", err
			}
			return fmt.Sprintf(`{"prim":"%s"}`, prim), nil
		}
		if ast.Nodes[i].Name != "expr" &&
			ast.Nodes[i].Name != "instrs" &&
			ast.Nodes[i].Name != "complex_instr" &&
			ast.Nodes[i].Name != "Int" &&
			ast.Nodes[i].Name != "String" &&
			ast.Nodes[i].Name != "Byte" {
			continue
		}

		return t.Translate(ast.Nodes[i])
	}
	return "", nil
}

func (t *MichelineTranslator) intTranslate(ast *peg.Ast) (string, error) {
	return fmt.Sprintf(`{"int":"%s"}`, ast.Token), nil
}

func (t *MichelineTranslator) stringTranslate(ast *peg.Ast) (string, error) {
	return fmt.Sprintf(`{"string":"%s"}`, sanitizeString(ast.Token)), nil
}

func (t *MichelineTranslator) bytesTranslate(ast *peg.Ast) (string, error) {
	return fmt.Sprintf(`{"bytes":"%s"}`, strings.TrimPrefix(ast.Token, "0x")), nil
}

func (t *MichelineTranslator) complexInstrTranslate(ast *peg.Ast) (string, error) {
	for i := range ast.Nodes {
		if ast.Nodes[i].Name != "instrs" {
			continue
		}

		return t.Translate(ast.Nodes[i])
	}
	return "[]", nil
}

func sanitizeString(token string) string {
	for from, to := range map[string]string{
		// "\\n": "\n",
		"\"": "",
	} {
		token = strings.ReplaceAll(token, from, to)
	}
	return token
}

func validatePrimitive(prim string) error {
	// TODO: handle macros + add FAIL prim
	valid, err := regexp.MatchString(
		"^parameter|storage|code|False|Elt|Left|None|Pair|Right|Some|True|Unit|PACK|UNPACK|BLAKE2B|SHA256|SHA512|ABS|ADD|AMOUNT|AND|BALANCE|CAR|CDR|CHECK_SIGNATURE|COMPARE|CONCAT|CONS|CREATE_CONTRACT|IMPLICIT_ACCOUNT|DIP|DROP|DUP|EDIV|EMPTY_MAP|EMPTY_SET|EQ|EXEC|FAILWITH|GE|GET|GT|HASH_KEY|IF|IF_CONS|IF_LEFT|IF_NONE|INT|LAMBDA|LE|LEFT|LOOP|LSL|LSR|LT|MAP|MEM|MUL|NEG|NEQ|NIL|NONE|NOT|NOW|OR|PAIR|PUSH|RIGHT|SIZE|SOME|SOURCE|SENDER|SELF|STEPS_TO_QUOTA|SUB|SWAP|TRANSFER_TOKENS|SET_DELEGATE|UNIT|UPDATE|XOR|ITER|LOOP_LEFT|ADDRESS|CONTRACT|ISNAT|CAST|RENAME|bool|contract|int|key|key_hash|lambda|list|map|big_map|nat|option|or|pair|set|signature|string|bytes|mutez|timestamp|unit|operation|address|SLICE|DIG|DUG|EMPTY_BIG_MAP|APPLY|chain_id|CHAIN_ID|LEVEL|SELF_ADDRESS|never|NEVER|UNPAIR|VOTING_POWER|TOTAL_VOTING_POWER|KECCAK|SHA3|PAIRING_CHECK|bls12_381_g1|bls12_381_g2|bls12_381_fr|sapling_state|sapling_transaction|SAPLING_EMPTY_STATE|SAPLING_VERIFY_UPDATE|ticket|TICKET|READ_TICKET|SPLIT_TICKET|JOIN_TICKETS|GET_AND_UPDATE|chest|chest_key|OPEN_CHEST|VIEW|view|constant$",
		prim)
	if err != nil {
		return err
	}
	if !valid {
		return errors.Errorf("Invalid primitive %s", prim)
	}
	return nil
}
