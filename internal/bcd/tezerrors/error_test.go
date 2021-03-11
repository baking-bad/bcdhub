package tezerrors

import (
	stdJSON "encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_parse(t *testing.T) {
	tests := []struct {
		name    string
		errJSON string
		ret     Error
	}{
		{
			name:    "Error 1",
			errJSON: "{\"kind\":\"temporary\",\"id\":\"proto.003-PsddFKi3.scriptRejectedRuntimeError\",\"location\":710,\"with\":{\"prim\":\"Unit\"}}",
			ret: Error{
				Kind:        "temporary",
				ID:          "proto.003-PsddFKi3.scriptRejectedRuntimeError",
				Title:       "Script failed (runtime script error)",
				Description: "A FAILWITH instruction was reached",
				IError: &DefaultError{
					Location: 710,
					With:     []byte(`{"prim":"Unit"}`),
				},
			},
		}, {
			name:    "Error 2",
			errJSON: "{\"kind\":\"temporary\",\"id\":\"proto.004-Pt24m4xi.gas_exhausted.operation\"}",
			ret: Error{
				Kind:        "temporary",
				ID:          "proto.004-Pt24m4xi.gas_exhausted.operation",
				Title:       "Gas quota exceeded for the operation",
				Description: "A script or one of its callee took more time than the operation said it would",
				IError:      &DefaultError{},
			},
		}, {
			name:    "Error 3",
			errJSON: "{\"kind\":\"temporary\",\"id\":\"proto.004-Pt24m4xi.contract.balance_too_low\",\"contract\":\"KT1BvVxWM6cjFuJNet4R9m64VDCN2iMvjuGE\",\"balance\":\"5248650175\",\"amount\":\"22571025048\"}",
			ret: Error{
				Kind:        "temporary",
				ID:          "proto.004-Pt24m4xi.contract.balance_too_low",
				Title:       "Balance too low",
				Description: "An operation tried to spend more tokens than the contract has",
				IError: &BalanceTooLowError{
					Amount:  22571025048,
					Balance: 5248650175,
				},
			},
		}, {
			name:    "Error 4",
			errJSON: "{\"kind\":\"temporary\",\"id\":\"proto.005-PsBabyM1.michelson_v1.script_rejected\",\"location\":226,\"with\":{\"prim\":\"Unit\"}}",
			ret: Error{
				Kind:        "temporary",
				ID:          "proto.005-PsBabyM1.michelson_v1.script_rejected",
				Title:       "Script failed",
				Description: "A FAILWITH instruction was reached",
				IError: &DefaultError{
					Location: 226,
					With:     []byte(`{"prim":"Unit"}`),
				},
			},
		}, {
			name:    "Error 5",
			errJSON: `{"kind": "permanent", "id": "proto.005-PsBabyM1.contract.manager.unregistered_delegate", "hash": "tz1YB12JHVHw9GbN66wyfakGYgdTBvokmXQk"}`,
			ret: Error{
				Kind:        "permanent",
				ID:          "proto.005-PsBabyM1.contract.manager.unregistered_delegate",
				Title:       "Unregistered delegate",
				Description: "A contract cannot be delegated to an unregistered delegate",
				IError:      &DefaultError{},
			},
		}, {
			name:    "Error 6",
			errJSON: `{ "kind": "temporary", "id": "proto.006-PsCARTHA.michelson_v1.script_rejected", "location": 1275, "with":{"string": "Wrong token type."}}`,
			ret: Error{
				Kind:        "temporary",
				ID:          "proto.006-PsCARTHA.michelson_v1.script_rejected",
				Title:       "Script failed",
				Description: "A FAILWITH instruction was reached",
				IError: &DefaultError{
					Location: 1275,
					With:     []byte(`{"string": "Wrong token type."}`),
				},
			},
		}, {
			name:    "Error 7",
			errJSON: `{"id":"proto.006-PsCARTHA.michelson_v1.script_rejected","kind":"temporary","location":3841,"with":{"prim":"Pair","args":[{"string":"AddrIsReg"},{"bytes":"0000e904e17b7f7f6b5456579b19b2ca0c96d9f31762"}]}}`,
			ret: Error{
				Kind:        "temporary",
				ID:          "proto.006-PsCARTHA.michelson_v1.script_rejected",
				Title:       "Script failed",
				Description: "A FAILWITH instruction was reached",
				IError: &DefaultError{
					Location: 3841,
					With:     []byte(`{"prim":"Pair","args":[{"string":"AddrIsReg"},{"bytes":"0000e904e17b7f7f6b5456579b19b2ca0c96d9f31762"}]}`),
				},
			},
		},
	}

	if err := LoadErrorDescriptions(); err != nil {
		panic(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			var e Error
			if err := json.Unmarshal([]byte(tt.errJSON), &e); err != nil {
				t.Errorf("json.Unmarshal: %v", err)
				return
			}

			if assert.NotNil(e) && assert.NotNil(tt.ret) {
				assert.Equalf(e, tt.ret, "Invalid parsed error: %##v != %##v", e.IError, tt.ret.IError)
			}
		})
	}
}

func TestBalanceTooLowError_Parse(t *testing.T) {
	tests := []struct {
		name string
		args string
		ret  Error
	}{
		{
			name: "Error 1",
			args: "{\"kind\":\"temporary\",\"id\":\"proto.004-Pt24m4xi.contract.balance_too_low\",\"contract\":\"KT1BvVxWM6cjFuJNet4R9m64VDCN2iMvjuGE\",\"balance\":\"5248650175\",\"amount\":\"22571025048\"}",
			ret: Error{
				Kind:        "temporary",
				ID:          "proto.004-Pt24m4xi.contract.balance_too_low",
				Title:       "Balance too low",
				Description: "An operation tried to spend more tokens than the contract has",
				IError: &BalanceTooLowError{
					Balance: 5248650175,
					Amount:  22571025048,
				},
			},
		},
	}

	if err := LoadErrorDescriptions(); err != nil {
		panic(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e Error
			if err := json.Unmarshal([]byte(tt.args), &e); err != nil {
				t.Errorf("json.Unmarshal: %v", err)
				return
			}

			if !reflect.DeepEqual(e, tt.ret) {
				t.Errorf("Invalid parsed error: %v != %v", e, tt.ret)
			}
		})
	}
}

func TestError_Format(t *testing.T) {
	tests := []struct {
		name        string
		args        *Error
		compareWith stdJSON.RawMessage
	}{
		{

			name: "Error 1",
			args: &Error{
				Kind:        "temporary",
				ID:          "proto.003-PsddFKi3.scriptRejectedRuntimeError",
				Title:       "Script failed (runtime script error)",
				Description: "A FAILWITH instruction was reached",
				IError: &DefaultError{
					Location: 710,
					With:     []byte(`{"prim":"Unit"}`),
				},
			},
			compareWith: []byte("Unit"),
		}, {
			name: "Error 2",
			args: &Error{
				Kind:        "temporary",
				ID:          "proto.004-Pt24m4xi.gas_exhausted.operation",
				Title:       "Gas quota exceeded for the operation",
				Description: "A script or one of its callee took more time than the operation said it would",
			},
		}, {
			name: "Error 3",
			args: &Error{
				Kind:        "temporary",
				ID:          "proto.004-Pt24m4xi.contract.balance_too_low",
				Title:       "Balance too low",
				Description: "An operation tried to spend more tokens than the contract has",
			},
		}, {
			name: "Error 4",
			args: &Error{
				Kind:        "temporary",
				ID:          "proto.005-PsBabyM1.michelson_v1.script_rejected",
				Title:       "Script failed",
				Description: "A FAILWITH instruction was reached",
				IError: &DefaultError{
					Location: 226,
					With:     []byte(`{"prim":"Unit"}`),
				},
			},
			compareWith: []byte("Unit"),
		}, {
			name: "Error 5",
			args: &Error{
				Kind:        "permanent",
				ID:          "proto.005-PsBabyM1.contract.manager.unregistered_delegate",
				Title:       "Unregistered delegate",
				Description: "A contract cannot be delegated to an unregistered delegate",
			},
		}, {
			name: "Error 6",
			args: &Error{
				Kind:        "temporary",
				ID:          "proto.004-Pt24m4xi.contract.balance_too_low",
				Title:       "Balance too low",
				Description: "An operation tried to spend more tokens than the contract has",
				IError: &BalanceTooLowError{
					Balance: 5248650175,
					Amount:  22571025048,
				},
			},
		}, {
			name: "Error 7",
			args: &Error{
				Kind:        "permanent",
				ID:          "proto.005-PsBabyM1.invalidSyntacticConstantError",
				Title:       "Invalid constant (parse error)",
				Description: "A compile-time constant was invalid for its expected form.",
				IError: &InvalidSyntacticConstantError{
					WrongExpressionSnake: []byte(`{"string":"KT1Mfe3rRhQw9KnEUZzoxkhmyHXBeN3zCzXL"}`),
					ExpectedFormSnake:    []byte(`{"prim":"key_hash"}`),
				},
			},
		},
	}

	if err := LoadErrorDescriptions(); err != nil {
		panic(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			err := tt.args.Format()
			if err != nil {
				t.Errorf("args format error %v", err)
				return
			}
			switch err := tt.args.IError.(type) {
			case *BalanceTooLowError:
				assert.Equalf(err.With, tt.compareWith, "Invalid formatted with error: %v != %v", err.With, tt.compareWith)
			case *DefaultError:
				assert.Equalf(err.With, tt.compareWith, "Invalid formatted with error: %v != %v", err.With, tt.compareWith)
			}
		})
	}
}

func TestInvalidSyntacticConstantError_Parse(t *testing.T) {
	tests := []struct {
		name string
		args string
		ret  Error
	}{
		{
			name: "Error 1",
			args: `{ "kind": "permanent", "id": "proto.005-PsBabyM1.invalidSyntacticConstantError", "location": 0, "expectedForm":{"prim": "unit"}, "wrongExpression":{"int": "0"}}`,
			ret: Error{
				Kind:        "permanent",
				ID:          "proto.005-PsBabyM1.invalidSyntacticConstantError",
				Title:       "Invalid constant (parse error)",
				Description: "A compile-time constant was invalid for its expected form.",
				IError: &InvalidSyntacticConstantError{
					ExpectedFormCamel:    []byte(`{"prim": "unit"}`),
					WrongExpressionCamel: []byte(`{"int": "0"}`),
				},
			},
		},
	}

	if err := LoadErrorDescriptions(); err != nil {
		panic(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e Error
			if err := json.Unmarshal([]byte(tt.args), &e); err != nil {
				t.Errorf("json.Unmarshal: %v", err)
				return
			}

			if !reflect.DeepEqual(e, tt.ret) {
				t.Errorf("Invalid parsed error: %v != %v", e, tt.ret)
			}
		})
	}
}
