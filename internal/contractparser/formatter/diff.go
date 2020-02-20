package formatter

import (
	"fmt"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/tidwall/gjson"
)

// DiffResult -
type DiffResult struct {
	DiffA   [][]Item `json:"diff_a"`
	DiffB   [][]Item `json:"diff_b"`
	NameA   string   `json:"name_a,omitempty"`
	NameB   string   `json:"name_b,omitempty"`
	Removed int64    `json:"removed"`
	Added   int64    `json:"added"`
}

// Item -
type Item struct {
	Type  int    `json:"type"`
	Chunk string `json:"chunk"`
}

// Diff -
func Diff(a, b gjson.Result) (res DiffResult, err error) {
	aString, err := MichelineToMichelson(a, true)
	if err != nil {
		return
	}

	bString, err := MichelineToMichelson(b, true)
	if err != nil {
		return
	}

	dmp := diffmatchpatch.New()
	diffList := dmp.DiffMain(aString, bString, true)

	aCode, err := MichelineToMichelson(a, false)
	if err != nil {
		return
	}

	bCode, err := MichelineToMichelson(b, false)
	if err != nil {
		return
	}
	diffA, err := diffToLines(aCode, diffList, -1)
	if err != nil {
		return
	}
	diffB, err := diffToLines(bCode, diffList, 1)
	if err != nil {
		return
	}

	return finish(diffA, diffB)
}

func diffToLines(text string, diff []diffmatchpatch.Diff, side int8) ([][]Item, error) {
	var ptr int
	res := make([][]Item, 0)
	line := make([]Item, 0)

	for i := range diff {
		if diff[i].Type != 0 && int8(diff[i].Type) != side {
			continue
		}

		var chunkOffset int
		chunkLen := len(diff[i].Text)

		for chunkLen > chunkOffset {
			if len(line) == 0 {
				chunkOffset = skipSpaces(diff[i].Text, chunkOffset)
				nextPtr := skipSpaces(text, ptr)
				if nextPtr > ptr {
					line = append(line, Item{
						Type:  int(side) * 2,
						Chunk: strings.Repeat(" ", nextPtr-ptr),
					})
					ptr = nextPtr
				}
			}

			nextOffset := chunkLen
			for j := chunkOffset; j < chunkLen; j++ {
				if text[ptr+j-chunkOffset] != diff[i].Text[j] {
					nextOffset = j
					break
				}
			}

			if nextOffset > chunkOffset {
				line = append(line, Item{
					Type:  int(diff[i].Type),
					Chunk: diff[i].Text[chunkOffset:nextOffset],
				})
				ptr += nextOffset - chunkOffset
				chunkOffset = nextOffset
			}

			if nextOffset < chunkLen {
				if text[ptr] != '\n' {
					return nil, fmt.Errorf("It is not end of line: %s, %d", text, ptr)
				}
				ptr++
				res = append(res, line)
				line = make([]Item, 0)
			}
		}
	}

	if len(line) != 0 {
		res = append(res, line)
	}

	return res, nil
}

func skipSpaces(s string, offset int) int {
	for i := offset; i < len(s); i++ {
		if s[i] != ' ' {
			return i
		}
	}
	return len(s)
}

func sign(i int) int {
	if i > 0 {
		return 1
	} else if i < 0 {
		return -1
	}
	return 0
}

func getLineSide(res [][]Item, i int) ([]Item, int) {
	if i < len(res) {
		var sum int
		for _, item := range res[i] {
			if item.Type == 0 {
				return res[i], 0
			}
			sum += item.Type
		}
		return res[i], sign(sum)
	}
	return []Item{}, 0
}

func finish(resA, resB [][]Item) (DiffResult, error) {
	res := DiffResult{
		DiffA: make([][]Item, 0),
		DiffB: make([][]Item, 0),
	}

	for a, b := 0, 0; a < len(resA) && b < len(resB); {
		lineA, sideA := getLineSide(resA, a)
		lineB, sideB := getLineSide(resB, b)

		if sideA == sideB {
			if sideA != 0 {
				return res, fmt.Errorf("Invalid side values [left] %d [right] %d", sideA, sideB)
			}
			res.DiffA = append(res.DiffA, lineA)
			res.DiffB = append(res.DiffB, lineB)
			a++
			b++
		} else if sideA == -1 {
			res.DiffA = append(res.DiffA, lineA)
			res.DiffB = append(res.DiffB, []Item{})
			res.Removed++
			a++
		} else if sideB == 1 {
			res.DiffA = append(res.DiffA, []Item{})
			res.DiffB = append(res.DiffB, lineB)
			res.Added++
			b++
		} else {
			return res, fmt.Errorf("Invalid side values [left] %d [right] %d", sideA, sideB)
		}
	}

	return res, nil
}
