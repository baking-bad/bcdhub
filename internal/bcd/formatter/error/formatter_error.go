package formattererror

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

type code struct {
	searchNode  int
	currentNode int
}

// LocateContractError - returns error position by number of error node
// returned values are: row, startCol, endCol, error
func LocateContractError(n gjson.Result, node int) (int, int, int, error) {
	c := &code{
		searchNode: node + 1,
	}

	text, err := c.locateError(n, "", true, false)
	if err != nil {
		return 0, 0, 0, err
	}

	row, startCol, endCol := findError(text)

	return row, startCol, endCol, nil
}

func (c *code) locateError(node gjson.Result, indent string, isRoot, wrapped bool) (string, error) {
	c.currentNode++

	if node.IsArray() {
		return c.locateInArray(node, indent, isRoot)
	}

	if node.IsObject() {
		return c.locateInObject(node, indent, isRoot, wrapped)
	}

	return "", errors.Errorf("node is not array or object: %v", node)
}

func (c *code) locateInArray(node gjson.Result, indent string, isRoot bool) (string, error) {
	seqIndent := indent
	isScriptRoot := isRoot && formatter.IsScript(node)
	if !isScriptRoot {
		seqIndent = indent + "  "
	}

	items := make([]string, len(node.Array()))

	for i, n := range node.Array() {
		res, err := c.locateError(n, seqIndent, false, true)
		if err != nil {
			return "", err
		}
		items[i] = res
	}

	if len(items) == 0 {
		return "{}", nil
	}

	length := len(indent) + 4
	for _, i := range items {
		length += len(i)
	}

	space := ""
	if !isScriptRoot {
		space = " "
	}

	var seq string

	if length < formatter.DefLineSize {
		seq = strings.Join(items, fmt.Sprintf("%v; ", space))
	} else {
		seq = strings.Join(items, fmt.Sprintf("%v;\n%v", space, seqIndent))
	}

	if !isScriptRoot {
		return fmt.Sprintf("{ %v }", seq), nil
	}

	return seq, nil
}

func (c *code) locateInObject(node gjson.Result, indent string, isRoot, wrapped bool) (string, error) {
	if node.Get("prim").Exists() {
		return c.locatePrimObject(node, indent, isRoot, wrapped)
	}

	return c.locateNonPrimObject(node)
}

func (c *code) locatePrimObject(node gjson.Result, indent string, isRoot, wrapped bool) (string, error) {
	var res []string

	prim := node.Get("prim").String()

	if c.searchNode == c.currentNode {
		res = append(res, unicodeMark(prim))
	} else {
		res = append(res, prim)
	}

	if annots := node.Get("annots"); annots.Exists() {
		for _, a := range annots.Array() {
			res = append(res, a.String())
		}
	}

	expr := strings.Join(res, " ")

	var args []gjson.Result
	if rawArgs := node.Get("args"); rawArgs.Exists() {
		args = rawArgs.Array()
	}

	switch {
	case formatter.IsComplex(node):
		argIndent := indent + "  "
		items := make([]string, len(args))
		for i, a := range args {
			res, err := c.locateError(a, argIndent, false, false)
			if err != nil {
				return "", err
			}
			items[i] = res
		}

		length := len(indent) + len(expr) + len(items) + 1

		for _, item := range items {
			length += len(item)
		}

		if length < formatter.DefLineSize {
			expr = fmt.Sprintf("%v %v", expr, strings.Join(items, " "))
		} else {
			res := []string{expr}
			res = append(res, items...)
			expr = strings.Join(res, fmt.Sprintf("\n%v", argIndent))
		}
	case len(args) == 1:
		argIndent := indent + strings.Repeat(" ", len(expr)+1)
		res, err := c.locateError(args[0], argIndent, false, false)
		if err != nil {
			return "", err
		}
		expr = fmt.Sprintf("%v %v", expr, res)
	case len(args) > 1:
		argIndent := indent + "  "
		altIndent := indent + strings.Repeat(" ", len(expr)+2)

		for _, arg := range args {
			item, err := c.locateError(arg, argIndent, false, false)
			if err != nil {
				return "", err
			}
			length := len(indent) + len(expr) + len(item) + 1
			if formatter.IsInline(node) || length < formatter.DefLineSize {
				argIndent = altIndent
				expr = fmt.Sprintf("%v %v", expr, item)
			} else {
				expr = fmt.Sprintf("%v\n%v%v", expr, argIndent, item)
			}
		}
	}

	if formatter.IsFramed(node) && !isRoot && !wrapped {
		return fmt.Sprintf("(%v)", expr), nil
	}
	return expr, nil
}

func (c *code) locateNonPrimObject(node gjson.Result) (string, error) {
	if len(node.Map()) != 1 {
		return "", errors.Errorf("node keys count != 1 %v", node)
	}

	for coreType, value := range node.Map() {
		switch coreType {
		case "int":
			return value.String(), nil
		case "bytes":
			return fmt.Sprintf("0x%v", value.String()), nil
		case "string":
			return value.Raw, nil
		}
	}

	return "", errors.Errorf("invalid coreType %v", node)
}

func unicodeMark(s string) string {
	if len(s) < 1 {
		return ""
	}

	return string(rune(s[0])+128) + s[1:]
}

func findError(text string) (int, int, int) {
	var found bool
	var row, start, end int

	rows := strings.Split(text, "\n")

	for i, r := range rows {
		for idx, c := range r {
			if isUnicode(c) {
				found = true
				row = i
				start = idx
				continue
			}

			if found && string(c) == " " {
				end = idx - 1
				return row, start, end
			}
		}
	}

	return 0, 0, 0
}

func isUnicode(r rune) bool {
	return r > unicode.MaxASCII
}
