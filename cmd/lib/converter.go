package main

import (
	"C"
	"encoding/json"
)

// Result -
type Result struct {
	Data  string `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// JSONConverter -
type JSONConverter struct{}

// Encode -
func (c JSONConverter) Encode(value interface{}) *C.char {
	result := Result{}
	data, err := json.Marshal(value)
	if err != nil {
		result.Error = err.Error()
	} else {
		result.Data = string(data)
	}
	response, _ := json.Marshal(result)
	return C.CString(string(response))
}

// EncodeBytes -
func (c JSONConverter) EncodeBytes(value []byte) *C.char {
	result := Result{
		Data: string(value),
	}
	response, _ := json.Marshal(result)
	return C.CString(string(response))
}

// EncodeError -
func (c JSONConverter) EncodeError(err error) *C.char {
	result := Result{Error: err.Error()}
	response, _ := json.Marshal(result)
	return C.CString(string(response))
}

// Decode -
func (c JSONConverter) Decode(data []byte, value interface{}) error {
	return json.Unmarshal(data, value)
}
