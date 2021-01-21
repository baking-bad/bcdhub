package search

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Scorable -
type Scorable interface {
	GetFields() []string
	GetIndex() string
	GetScores(string) []string
	Parse(highlight map[string][]string, data []byte) (interface{}, error)
}

// Indices - list of indices availiable to search
var Indices = []string{
	models.DocContracts,
	models.DocOperations,
	models.DocBigMapDiff,
	models.DocTokenMetadata,
	models.DocTZIP,
	models.DocTezosDomains,
}

var scorables = map[string]Scorable{
	models.DocContracts:     &Contract{},
	models.DocOperations:    &Operation{},
	models.DocBigMapDiff:    &BigMap{},
	models.DocTokenMetadata: &Token{},
	models.DocTZIP:          &Metadata{},
	models.DocTezosDomains:  &Domain{},
}

// ScoreInfo -
type ScoreInfo struct {
	Scores  []string
	Indices []string

	indicesMap map[string]struct{}
	fieldsMap  map[string]int
}

func newScoreInfo() ScoreInfo {
	return ScoreInfo{
		Scores:  make([]string, 0),
		Indices: make([]string, 0),

		indicesMap: make(map[string]struct{}),
		fieldsMap:  make(map[string]int),
	}
}

func (si *ScoreInfo) addIndex(index string) {
	if _, ok := si.indicesMap[index]; ok {
		return
	}
	si.indicesMap[index] = struct{}{}
	si.Indices = append(si.Indices, index)
}

func (si *ScoreInfo) addScore(score string) error {
	val := strings.Split(score, "^")
	field := val[0]

	var iScore int
	if len(val) == 1 {
		iScore = 1
	} else if len(val) == 2 {
		i, err := strconv.Atoi(val[1])
		if err != nil {
			return err
		}
		iScore = i
	}

	j, ok := si.fieldsMap[field]
	if !ok || j < iScore {
		si.fieldsMap[field] = iScore
	}

	return nil
}

func (si *ScoreInfo) addScores(scores ...string) error {
	for i := range scores {
		if err := si.addScore(scores[i]); err != nil {
			return err
		}
	}

	return nil
}

// GetScores -
func GetScores(searchString string, fields []string, indices ...string) (ScoreInfo, error) {
	info := newScoreInfo()
	if len(indices) == 0 {
		indices = Indices
	}

	for i := range indices {
		model, ok := scorables[indices[i]]
		if !ok {
			return info, errors.Errorf("[GetSearchScores] Unknown scorable model: %s", indices[i])
		}
		index := model.GetIndex()
		if helpers.StringInArray(index, Indices) {
			modelScores := model.GetScores(searchString)
			info.addIndex(index)
			if len(fields) > 0 {
				for i := range modelScores {
					for j := range fields {
						if strings.HasPrefix(modelScores[i], fields[j]) {
							if err := info.addScore(modelScores[i]); err != nil {
								return info, err
							}
						}
					}
				}
			} else {
				if err := info.addScores(modelScores...); err != nil {
					return info, err
				}
			}
		}
	}

	for k, v := range info.fieldsMap {
		info.Scores = append(info.Scores, fmt.Sprintf("%s^%d", k, v))
	}

	return info, nil
}

// Parse -
func Parse(index string, highlight map[string][]string, data []byte) (interface{}, error) {
	if s, ok := scorables[index]; ok {
		return s.Parse(highlight, data)
	}
	return nil, errors.Errorf("Unknown index: %s", index)
}
