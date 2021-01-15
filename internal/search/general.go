package search

import (
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
	fieldsMap  map[string]struct{}
}

func newScoreInfo() ScoreInfo {
	return ScoreInfo{
		Scores:  make([]string, 0),
		Indices: make([]string, 0),

		indicesMap: make(map[string]struct{}),
		fieldsMap:  make(map[string]struct{}),
	}
}

func (si *ScoreInfo) addIndex(index string) {
	if _, ok := si.indicesMap[index]; ok {
		return
	}
	si.indicesMap[index] = struct{}{}
	si.Indices = append(si.Indices, index)
}

func (si *ScoreInfo) addScore(score string) {
	val := strings.Split(score, "^")
	field := val[0]
	if _, ok := si.fieldsMap[field]; ok {
		return
	}
	si.fieldsMap[field] = struct{}{}
	si.Scores = append(si.Scores, score)
}

func (si *ScoreInfo) addScores(scores ...string) {
	for i := range scores {
		si.addScore(scores[i])
	}
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
							info.addScore(modelScores[i])
						}
					}
				}
			} else {
				info.addScores(modelScores...)
			}
		}
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
