package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/macros"
	"github.com/aopoltorzhicky/bcdhub/internal/helpers"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

func computeFingerprint(script gjson.Result, contract *models.Contract) error {
	colapsed, err := macros.FindMacros(script)
	if err != nil {
		return err
	}

	fgpt := models.Fingerprint{}
	code := colapsed.Get(`code.#(prim="code")`)
	codeFgpt, err := fingerprint(code, true)
	if err != nil {
		return err
	}
	fgpt.Code = codeFgpt

	params := colapsed.Get(`code.#(prim="parameter")`)
	paramFgpt, err := fingerprint(params, false)
	if err != nil {
		return err
	}
	fgpt.Parameter = paramFgpt

	storage := colapsed.Get(`code.#(prim="storage")`)
	storageFgpt, err := fingerprint(storage, false)
	if err != nil {
		return err
	}
	fgpt.Storage = storageFgpt

	contract.Fingerprint = &fgpt
	return nil
}

func fingerprint(script gjson.Result, isCode bool) (string, error) {
	var fgpt strings.Builder
	if script.IsObject() {
		prim := script.Get(consts.KeyPrim)
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

			args := script.Get(consts.KeyArgs)
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

				if !contractparser.IsLiteral(k) {
					itemFgpt, err := fingerprint(v, isCode)
					if err != nil {
						return "", err
					}
					fgpt.WriteString(itemFgpt)
				}
			}
		}
	} else if script.IsArray() {
		for _, item := range script.Array() {
			buf, err := fingerprint(item, isCode)
			if err != nil {
				return "", err
			}
			fgpt.WriteString(buf)
		}
	} else {
		return "", fmt.Errorf("Unknwon script type: %v isCode: %v", script, isCode)
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
	return "00", fmt.Errorf("Unknown primitive: %s", prim)
}
