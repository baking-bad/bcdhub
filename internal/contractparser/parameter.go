package contractparser

import (
	"log"
	"strings"
)

func (c *Code) parseEntrypointArgs(args interface{}) []string {
	if args == nil {
		return nil
	}
	return c.argsToFlat(args)
}

func (c *Code) argsToFlat(args interface{}) []string {
	res := make([]string, 0)
	switch val := args.(type) {
	case []interface{}:
		for i := range val {
			res = append(res, c.argsToFlat(val[i])...)
		}
	case map[string]interface{}:
		prim, ok := val["prim"]
		if !ok {
			return nil
		}
		sPrim := strings.ToUpper(prim.(string))

		c.Primitives.Append(sPrim)

		args, ok := val["args"]
		if sPrim == "PAIR" || sPrim == "OR" {
			if ok {
				res = append(res, c.argsToFlat(args)...)
			}
		}

		if sPrim == "CONTRACT" {
			c.Tags.Append(ViewMethodTag)
		}

		if ok {
			res = append(res, c.argsToFlat(args)...)
		} else {
			res = append(res, sPrim)
		}
	}
	return res
}

func (c *Code) findEntrypoint(v map[string]interface{}) {
	args, ok := v["args"]
	if !ok {
		return
	}
	annots, ok := v["annots"]
	if ok {
		for _, v := range annots.([]interface{}) {
			name := strings.Trim(v.(string), "%@:")
			c.Entrypoints[name] = c.parseEntrypointArgs(args)
		}
	}
	for _, a := range args.([]interface{}) {
		c.findEntrypoint(a.(map[string]interface{}))
	}
}

func (c *Code) print() {
	log.Print("Entrypoints:")
	for name, args := range c.Entrypoints {
		log.Printf("%s(%s)", name, strings.Join(args, ","))
	}
}
