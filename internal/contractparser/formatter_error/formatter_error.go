package formattererror

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/formatter"
	"github.com/tidwall/gjson"
)

type code struct {
	searchNode  int
	currentNode int
}

// LocateContractError - returns error position by number of error node
// returned values are: row, startCol, endCol
func LocateContractError(n gjson.Result, node int) (int, int, int) {
	c := &code{
		searchNode: node + 1,
	}

	text := c.locateError(n, "", true, false)

	return findError(text)
}

func (c *code) locateError(node gjson.Result, indent string, isRoot, wrapped bool) string {
	c.currentNode++

	if node.IsArray() {
		return c.locateInArray(node, indent, isRoot)
	}

	if node.IsObject() {
		return c.locateInObject(node, indent, isRoot, wrapped)
	}

	fmt.Println("NODE:", node)
	panic("shit happens")
}

func (c *code) locateInArray(node gjson.Result, indent string, isRoot bool) string {
	seqIndent := indent
	isScriptRoot := isRoot && formatter.IsScript(node)
	if !isScriptRoot {
		seqIndent = indent + "  "
	}

	items := make([]string, len(node.Array()))

	for i, n := range node.Array() {
		items[i] = c.locateError(n, seqIndent, false, true)
	}

	if len(items) == 0 {
		return "{}"
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

	if length < formatter.LineSize {
		seq = strings.Join(items, fmt.Sprintf("%v; ", space))
	} else {
		seq = strings.Join(items, fmt.Sprintf("%v;\n%v", space, seqIndent))
	}

	if !isScriptRoot {
		return fmt.Sprintf("{ %v }", seq)
	}

	return seq
}

func (c *code) locateInObject(node gjson.Result, indent string, isRoot, wrapped bool) string {
	if node.Get("prim").Exists() {
		return c.locatePrimObject(node, indent, isRoot, wrapped)
	}

	return c.locateNonPrimObject(node)
}

func (c *code) locatePrimObject(node gjson.Result, indent string, isRoot, wrapped bool) string {
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

	if formatter.IsComplex(node) {
		argIndent := indent + "  "
		items := make([]string, len(args))
		for i, a := range args {
			items[i] = c.locateError(a, argIndent, false, false)
		}

		length := len(indent) + len(expr) + len(items) + 1

		for _, item := range items {
			length += len(item)
		}

		if length < formatter.LineSize {
			expr = fmt.Sprintf("%v %v", expr, strings.Join(items, " "))
		} else {
			res := []string{expr}
			res = append(res, items...)
			expr = strings.Join(res, fmt.Sprintf("\n%v", argIndent))
		}
	} else if len(args) == 1 {
		argIndent := indent + strings.Repeat(" ", len(expr)+1)
		expr = fmt.Sprintf("%v %v", expr, c.locateError(args[0], argIndent, false, false))
	} else if len(args) > 1 {
		argIndent := indent + "  "
		altIndent := indent + strings.Repeat(" ", len(expr)+2)

		for _, arg := range args {
			item := c.locateError(arg, argIndent, false, false)
			length := len(indent) + len(expr) + len(item) + 1
			if formatter.IsInline(node) || length < formatter.LineSize {
				argIndent = altIndent
				expr = fmt.Sprintf("%v %v", expr, item)
			} else {
				expr = fmt.Sprintf("%v\n%v%v", expr, argIndent, item)
			}
		}
	}

	if formatter.IsFramed(node) && !isRoot && !wrapped {
		return fmt.Sprintf("(%v)", expr)
	}
	return expr
}

func (c *code) locateNonPrimObject(node gjson.Result) string {
	if len(node.Map()) != 1 {
		fmt.Println("NODE:", node)
		panic("node keys count != 1")
	}

	for coreType, value := range node.Map() {
		if coreType == "int" {
			return value.String()
		} else if coreType == "bytes" {
			return fmt.Sprintf("0x%v", value.String())
		} else if coreType == "string" {
			return value.Raw
		}
	}

	fmt.Println("NODE:", node)
	panic("invalid coreType")
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
