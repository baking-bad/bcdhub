package cerrors

import (
	"reflect"
	"testing"

	"github.com/tidwall/gjson"
)

func TestError_parse(t *testing.T) {
	tests := []struct {
		name string
		args string
		ret  Error
	}{
		{
			name: "Error 1",
			args: "{\"kind\":\"temporary\",\"id\":\"proto.003-PsddFKi3.scriptRejectedRuntimeError\",\"location\":710,\"with\":{\"prim\":\"Unit\"}}",
			ret: Error{
				Kind:        "temporary",
				ID:          "proto.003-PsddFKi3.scriptRejectedRuntimeError",
				Location:    710,
				With:        "{\"prim\":\"Unit\"}",
				Title:       "Script failed (runtime script error)",
				Description: "A FAILWITH instruction was reached",
			},
		}, {
			name: "Error 2",
			args: "{\"kind\":\"temporary\",\"id\":\"proto.004-Pt24m4xi.gas_exhausted.operation\"}",
			ret: Error{
				Kind:        "temporary",
				ID:          "proto.004-Pt24m4xi.gas_exhausted.operation",
				Title:       "Gas quota exceeded for the operation",
				Description: "A script or one of its callee took more time than the operation said it would",
			},
		}, {
			name: "Error 3",
			args: "{\"kind\":\"temporary\",\"id\":\"proto.004-Pt24m4xi.contract.balance_too_low\",\"contract\":\"KT1BvVxWM6cjFuJNet4R9m64VDCN2iMvjuGE\",\"balance\":\"5248650175\",\"amount\":\"22571025048\"}",
			ret: Error{
				Kind:        "temporary",
				ID:          "proto.004-Pt24m4xi.contract.balance_too_low",
				Title:       "Balance too low",
				Description: "An operation tried to spend more tokens than the contract has",
			},
		}, {
			name: "Error 4",
			args: "{\"kind\":\"temporary\",\"id\":\"proto.005-PsBabyM1.michelson_v1.script_rejected\",\"location\":226,\"with\":{\"prim\":\"Unit\"}}",
			ret: Error{
				Kind:        "temporary",
				ID:          "proto.005-PsBabyM1.michelson_v1.script_rejected",
				Title:       "Script failed",
				Description: "A FAILWITH instruction was reached",
				Location:    226,
				With:        "{\"prim\":\"Unit\"}",
			},
		},
	}

	if err := loadErrorDescriptions(); err != nil {
		panic(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e Error

			data := gjson.Parse(tt.args)
			e.Parse(data)

			if !reflect.DeepEqual(e, tt.ret) {
				t.Errorf("Invalid parsed error: %v != %v", e, tt.ret)
			}
		})
	}
}
