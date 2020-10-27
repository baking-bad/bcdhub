package formatter

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/tidwall/gjson"
)

// DiffLineSize -
const DiffLineSize = 74

// DiffResult -
type DiffResult struct {
	Left    [][]Item `json:"left"`
	Right   [][]Item `json:"right"`
	Removed int64    `json:"removed"`
	Added   int64    `json:"added"`
	Changed int64    `json:"changed"`
}

// Item -
type Item struct {
	Type  int    `json:"type"`
	Chunk string `json:"chunk"`
	ID    int    `json:"-"`
}

// Compare -
func (i Item) Compare(other Item) bool {
	return i.Type == other.Type && i.ID == other.ID
}

// Diff -
func Diff(a, b gjson.Result) (res DiffResult, err error) {
	aString, err := MichelineToMichelson(a, true, DiffLineSize)
	if err != nil {
		return
	}

	bString, err := MichelineToMichelson(b, true, DiffLineSize)
	if err != nil {
		return
	}

	dmp := diffmatchpatch.New()
	diffList := dmp.DiffMain(aString, bString, true)
	diffList = dmp.DiffCleanupSemantic(diffList)

	aCode, err := MichelineToMichelson(a, false, DiffLineSize)
	if err != nil {
		return
	}

	bCode, err := MichelineToMichelson(b, false, DiffLineSize)
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

	return postProcessing(diffA, diffB), nil
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
						ID:    i,
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
					ID:    i,
				})
				ptr += nextOffset - chunkOffset
				chunkOffset = nextOffset
			}

			if nextOffset < chunkLen {
				if text[ptr] != '\n' {
					return nil, errors.Errorf("It is not end of line: %s, %d", text, ptr)
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

//nolint
func postProcessing(resA, resB [][]Item) DiffResult {
	res := DiffResult{
		Left:  make([][]Item, 0),
		Right: make([][]Item, 0),
	}

	var lIdx, rIdx int
	for l, r := 0, 0; l < len(resA) || r < len(resB); {
		newLeft := make([]Item, 0)
		newRight := make([]Item, 0)

		if l == len(resA) {
			right := resB[r]
			for i := rIdx; i < len(right); i++ {
				newRight = append(newRight, right[i])
			}
			res.Right = append(res.Right, newRight)
			res.Left = append(res.Left, []Item{})
			r++
			rIdx = 0
			continue
		}
		if r == len(resB) {
			left := resA[l]
			for i := lIdx; i < len(left); i++ {
				newLeft = append(newLeft, left[i])
			}
			res.Left = append(res.Left, newLeft)
			res.Right = append(res.Right, []Item{})
			l++
			lIdx = 0
			continue
		}

		left := resA[l]
		right := resB[r]
		leftType := 0
		rightType := 0

		for lIdx < len(left) || rIdx < len(right) {
			if lIdx == len(left) {
				if right[rIdx].Type == 0 {
					break
				}
				if right[rIdx].Type > -2 && right[rIdx].Type < 2 {
					right[rIdx].Type = 1
					rightType = 1
				}
				newRight = append(newRight, right[rIdx])
				rIdx++
				continue
			} else if rIdx == len(right) {
				if left[lIdx].Type == 0 {
					break
				}
				if left[lIdx].Type > -2 && left[lIdx].Type < 2 {
					left[lIdx].Type = -1
					leftType = -1
				}
				newLeft = append(newLeft, left[lIdx])
				lIdx++
				continue
			}

			if left[lIdx].Compare(right[rIdx]) {
				newLeft = append(newLeft, left[lIdx])
				newRight = append(newRight, right[rIdx])
				lIdx++
				rIdx++
				continue
			}

			if left[lIdx].ID > right[rIdx].ID || left[lIdx].Type == 0 {
				if right[rIdx].Type > -2 && right[rIdx].Type < 2 {
					right[rIdx].Type = 1
					rightType = 1
				}
				newRight = append(newRight, right[rIdx])
				rIdx++
				continue
			} else if left[lIdx].ID < right[rIdx].ID || right[rIdx].Type == 0 {
				if left[lIdx].Type > -2 && left[lIdx].Type < 2 {
					left[lIdx].Type = -1
					leftType = -1
				}
				newLeft = append(newLeft, left[lIdx])
				lIdx++
				continue
			} else {
				newLeft = append(newLeft, left[lIdx])
				newRight = append(newRight, right[rIdx])
				lIdx++
				rIdx++
				continue
			}
		}

		res.Left = append(res.Left, newLeft)
		res.Right = append(res.Right, newRight)
		if leftType == -1 {
			res.Removed++
		}
		if rightType == 1 {
			res.Added++
		}
		if lIdx == len(left) {
			lIdx = 0
			l++
		}
		if len(right) == rIdx {
			rIdx = 0
			r++
		}
	}

	return res
}
