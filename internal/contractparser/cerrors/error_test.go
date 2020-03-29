package cerrors

import (
	"reflect"
	"testing"

	"github.com/tidwall/gjson"
)

func TestDefaultError_parse(t *testing.T) {
	tests := []struct {
		name string
		args string
		ret  DefaultError
	}{
		{
			name: "Error 1",
			args: "{\"kind\":\"temporary\",\"id\":\"proto.003-PsddFKi3.scriptRejectedRuntimeError\",\"location\":710,\"with\":{\"prim\":\"Unit\"}}",
			ret: DefaultError{
				Kind:        "temporary",
				ID:          "proto.003-PsddFKi3.scriptRejectedRuntimeError",
				Title:       "Script failed (runtime script error)",
				Description: "A FAILWITH instruction was reached",
				Location:    710,
				With:        `{"prim":"Unit"}`,
			},
		}, {
			name: "Error 2",
			args: "{\"kind\":\"temporary\",\"id\":\"proto.004-Pt24m4xi.gas_exhausted.operation\"}",
			ret: DefaultError{
				Kind:        "temporary",
				ID:          "proto.004-Pt24m4xi.gas_exhausted.operation",
				Title:       "Gas quota exceeded for the operation",
				Description: "A script or one of its callee took more time than the operation said it would",
			},
		}, {
			name: "Error 3",
			args: "{\"kind\":\"temporary\",\"id\":\"proto.004-Pt24m4xi.contract.balance_too_low\",\"contract\":\"KT1BvVxWM6cjFuJNet4R9m64VDCN2iMvjuGE\",\"balance\":\"5248650175\",\"amount\":\"22571025048\"}",
			ret: DefaultError{
				Kind:        "temporary",
				ID:          "proto.004-Pt24m4xi.contract.balance_too_low",
				Title:       "Balance too low",
				Description: "An operation tried to spend more tokens than the contract has",
			},
		}, {
			name: "Error 4",
			args: "{\"kind\":\"temporary\",\"id\":\"proto.005-PsBabyM1.michelson_v1.script_rejected\",\"location\":226,\"with\":{\"prim\":\"Unit\"}}",
			ret: DefaultError{
				Kind:        "temporary",
				ID:          "proto.005-PsBabyM1.michelson_v1.script_rejected",
				Title:       "Script failed",
				Description: "A FAILWITH instruction was reached",
				Location:    226,
				With:        `{"prim":"Unit"}`,
			},
		}, {
			name: "Error 5",
			args: `{"kind": "permanent", "id": "proto.005-PsBabyM1.contract.manager.unregistered_delegate", "hash": "tz1YB12JHVHw9GbN66wyfakGYgdTBvokmXQk"}`,
			ret: DefaultError{
				Kind:        "permanent",
				ID:          "proto.005-PsBabyM1.contract.manager.unregistered_delegate",
				Title:       "Unregistered delegate",
				Description: "A contract cannot be delegated to an unregistered delegate",
			},
		},
	}

	if err := LoadErrorDescriptions("errors.json"); err != nil {
		panic(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e DefaultError

			data := gjson.Parse(tt.args)
			e.Parse(data)

			if !reflect.DeepEqual(e, tt.ret) {
				t.Errorf("Invalid parsed error: %v != %v", e, tt.ret)
			}
		})
	}
}

func TestBalanceTooLowError_Parse(t *testing.T) {
	tests := []struct {
		name string
		args string
		ret  BalanceTooLowError
	}{
		{
			name: "Error 1",
			args: "{\"kind\":\"temporary\",\"id\":\"proto.004-Pt24m4xi.contract.balance_too_low\",\"contract\":\"KT1BvVxWM6cjFuJNet4R9m64VDCN2iMvjuGE\",\"balance\":\"5248650175\",\"amount\":\"22571025048\"}",
			ret: BalanceTooLowError{
				DefaultError: DefaultError{
					Kind:        "temporary",
					ID:          "proto.004-Pt24m4xi.contract.balance_too_low",
					Title:       "Balance too low",
					Description: "An operation tried to spend more tokens than the contract has",
				},
				Balance: 5248650175,
				Amount:  22571025048,
			},
		},
	}

	if err := LoadErrorDescriptions("errors.json"); err != nil {
		panic(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e BalanceTooLowError

			data := gjson.Parse(tt.args)
			e.Parse(data)

			if !reflect.DeepEqual(e, tt.ret) {
				t.Errorf("Invalid parsed error: %v != %v", e, tt.ret)
			}
		})
	}
}

func TestDefaultError_Format(t *testing.T) {
	tests := []struct {
		name        string
		args        IError
		compareWith string
	}{
		{
			name: "Error 1",
			args: &DefaultError{
				Kind:        "temporary",
				ID:          "proto.003-PsddFKi3.scriptRejectedRuntimeError",
				Title:       "Script failed (runtime script error)",
				Description: "A FAILWITH instruction was reached",
				Location:    710,
				With:        `{"prim":"Unit"}`,
			},
			compareWith: "Unit",
		}, {
			name: "Error 2",
			args: &DefaultError{
				Kind:        "temporary",
				ID:          "proto.004-Pt24m4xi.gas_exhausted.operation",
				Title:       "Gas quota exceeded for the operation",
				Description: "A script or one of its callee took more time than the operation said it would",
			},
		}, {
			name: "Error 3",
			args: &DefaultError{
				Kind:        "temporary",
				ID:          "proto.004-Pt24m4xi.contract.balance_too_low",
				Title:       "Balance too low",
				Description: "An operation tried to spend more tokens than the contract has",
			},
		}, {
			name: "Error 4",
			args: &DefaultError{
				Kind:        "temporary",
				ID:          "proto.005-PsBabyM1.michelson_v1.script_rejected",
				Title:       "Script failed",
				Description: "A FAILWITH instruction was reached",
				Location:    226,
				With:        `{"prim":"Unit"}`,
			},
			compareWith: "Unit",
		}, {
			name: "Error 5",
			args: &DefaultError{
				Kind:        "permanent",
				ID:          "proto.005-PsBabyM1.contract.manager.unregistered_delegate",
				Title:       "Unregistered delegate",
				Description: "A contract cannot be delegated to an unregistered delegate",
			},
		}, {
			name: "Error 6",
			args: &BalanceTooLowError{
				DefaultError: DefaultError{
					Kind:        "temporary",
					ID:          "proto.004-Pt24m4xi.contract.balance_too_low",
					Title:       "Balance too low",
					Description: "An operation tried to spend more tokens than the contract has",
				},
				Balance: 5248650175,
				Amount:  22571025048,
			},
		},
	}

	if err := LoadErrorDescriptions("errors.json"); err != nil {
		panic(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.Format()
			switch err := tt.args.(type) {
			case *BalanceTooLowError:
				if err.With != tt.compareWith {
					t.Errorf("Invalid formatted with error: %v != %v", err.With, tt.compareWith)
				}
			case *DefaultError:
				if err.With != tt.compareWith {
					t.Errorf("Invalid formatted with error: %v != %v", err.With, tt.compareWith)
				}
			}
		})
	}
}
