package storage

import (
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

func getResult(op gjson.Result) (gjson.Result, error) {
	result := op.Get("metadata.operation_result")
	if !result.Exists() {
		result = op.Get("result")
		if !result.Exists() {
			return gjson.Result{}, errors.Errorf("[storage.getResult] Can not find 'result'")
		}
	}
	return result, nil
}
