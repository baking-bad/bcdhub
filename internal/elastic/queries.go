package elastic

import "encoding/json"

var (
	queryAll = map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"size": 10000,
	}
)

func createQueryMap(s string) (map[string]interface{}, error) {
	var res map[string]interface{}
	err := json.Unmarshal([]byte(s), &res)
	return res, err
}
