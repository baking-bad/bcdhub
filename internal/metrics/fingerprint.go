package metrics

import (
	"regexp"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/macros"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// SetFingerprint -
func SetFingerprint(script gjson.Result, contract *contract.Contract) error {
	fgpt, err := GetFingerprint(script)
	if err != nil {
		return err
	}
	contract.Fingerprint = fgpt
	return nil
}

// GetFingerprint -
func GetFingerprint(script gjson.Result) (*contract.Fingerprint, error) {
	colapsed, err := macros.Collapse(script, macros.GetAllFamilies())
	if err != nil {
		return nil, err
	}
	fgpt := contract.Fingerprint{}
	code := colapsed.Get(`code.#(prim="code")`)
	codeFgpt, err := fingerprint(code, true)
	if err != nil {
		return nil, err
	}
	fgpt.Code = codeFgpt

	params := colapsed.Get(`code.#(prim="parameter")`)
	paramFgpt, err := fingerprint(params, false)
	if err != nil {
		return nil, err
	}
	fgpt.Parameter = paramFgpt

	storage := colapsed.Get(`code.#(prim="storage")`)
	storageFgpt, err := fingerprint(storage, false)
	if err != nil {
		return nil, err
	}
	fgpt.Storage = storageFgpt
	return &fgpt, nil
}

func fingerprint(script gjson.Result, isCode bool) (string, error) {
	var fgpt strings.Builder
	switch {
	case script.IsObject():
		prim := script.Get("prim")
		if prim.Exists() {
			sPrim := prim.String()

			if skip(sPrim, isCode) {
				return "", nil
			}

			if !pass(sPrim, isCode) {
				code, err := getCode(sPrim)
				if err != nil {
					return "", err
				}
				fgpt.WriteString(code)
			}

			args := script.Get("args")
			if args.Exists() {
				itemFgpt, err := fingerprint(args, isCode)
				if err != nil {
					return "", err
				}
				fgpt.WriteString(itemFgpt)
			}

		} else {
			for k, v := range script.Map() {
				code, err := getCode(k)
				if err != nil {
					return "", err
				}
				fgpt.WriteString(code)

				if !bcd.IsLiteral(k) {
					itemFgpt, err := fingerprint(v, isCode)
					if err != nil {
						return "", err
					}
					fgpt.WriteString(itemFgpt)
				}
			}
		}
	case script.IsArray():
		for _, item := range script.Array() {
			buf, err := fingerprint(item, isCode)
			if err != nil {
				return "", err
			}
			fgpt.WriteString(buf)
		}
	default:
		return "", errors.Errorf("Unknown script type: %v isCode: %v", script, isCode)
	}

	return fgpt.String(), nil
}

func skip(prim string, isCode bool) bool {
	p := strings.ToLower(prim)
	return isCode && helpers.StringInArray(p, []string{
		consts.CAST, consts.RENAME,
	})
}

func pass(prim string, isCode bool) bool {
	p := strings.ToLower(prim)
	return !isCode && helpers.StringInArray(p, []string{
		consts.PAIR, consts.OR,
	})
}

func getCode(prim string) (string, error) {
	code, ok := codes[prim]
	if ok {
		return code, nil
	}

	for template, code := range regCodes {
		if template[0] != prim[0] {
			continue
		}
		re := regexp.MustCompile(template)
		if re.MatchString(prim) {
			return code, nil
		}
	}
	return "00", errors.Errorf("Unknown primitive: %s", prim)
}
