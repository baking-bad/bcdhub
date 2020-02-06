package storage

import (
	"fmt"

	"github.com/tidwall/gjson"
)

func getResult(op gjson.Result) (gjson.Result, error) {
	result := op.Get("metadata.operation_result")
	if !result.Exists() {
		result = op.Get("result")
		if !result.Exists() {
			return gjson.Result{}, fmt.Errorf("[storage.getResult] Can not find 'result'")
		}
	}
	return result, nil
}
