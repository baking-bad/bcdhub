package tzip

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/stretchr/testify/assert"
)

func Test_bufTzip_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    *bufTzip
		wantErr bool
	}{
		{
			name: "KT1CUwx4yV7dLjce6sh6ttrwHnAcBTNduw8j",
			data: `{"description":"A very Bazaar collection contract.","interfaces":["TZIP-012","TZIP-016","TZIP-020"],"tokenCategory":"collectibles","name":"The First Collection"}`,
			want: &bufTzip{
				TZIP16: tzip.TZIP16{
					Name:        "The First Collection",
					Description: "A very Bazaar collection contract.",
					Interfaces:  []string{"TZIP-012", "TZIP-016", "TZIP-020"},
				},
				Extras: map[string]interface{}{
					"tokenCategory": "collectibles",
				},
			},
		}, {
			name: "KT1CUwx4yV7dLjce6sh6ttrwHnAcBTNduw8j with license",
			data: `{"description":"A very Bazaar collection contract.","interfaces":["TZIP-012","TZIP-016","TZIP-020"],"tokenCategory":"collectibles","name":"The First Collection","license":"MIT"}`,
			want: &bufTzip{
				TZIP16: tzip.TZIP16{
					Name:        "The First Collection",
					Description: "A very Bazaar collection contract.",
					Interfaces:  []string{"TZIP-012", "TZIP-016", "TZIP-020"},
					License: &tzip.License{
						Name: "MIT",
					},
				},
				Extras: map[string]interface{}{
					"tokenCategory": "collectibles",
				},
			},
		}, {
			name: "KT1RVvbhbjvkpCtauT32aGiySWEKMYbbDfQS",
			data: `{ "version": "V0.0.1", "description": "This is based on a didactic reference implementation of FA2, a.k.a. TZIP-012, using SmartPy and template https://gitlab.com/smondet/fa2-smartpy.git.\n\nThis particular contract uses the configuration named: FA2.", "interfaces": [ "TZIP-012-2020-12-24", "TZIP-016" ], "authors": [ "MUSCADE" ], "homepage": "TBD", "source": { "tools": [ "SmartPy" ], "location": "TBD" }, "permissions": { "operator": "owner-or-operator-transfer", "receiver": "owner-no-hook", "sender": "owner-no-hook" }, "fa2-smartpy": { "configuration": { "add_mutez_transfer": false, "allow_self_transfer": false, "assume_consecutive_token_ids": true, "force_layouts": true, "lazy_entry_points": false, "lazy_entry_points_multiple": false, "name": "FA2", "non_fungible": false, "readable": true, "single_asset": false, "store_total_supply": true, "support_operator": true } } }`,
			want: &bufTzip{
				TZIP16: tzip.TZIP16{
					Version:     "V0.0.1",
					Description: "This is based on a didactic reference implementation of FA2, a.k.a. TZIP-012, using SmartPy and template https://gitlab.com/smondet/fa2-smartpy.git.\n\nThis particular contract uses the configuration named: FA2.",
					Interfaces:  []string{"TZIP-012-2020-12-24", "TZIP-016"},
					Authors:     []string{"MUSCADE"},
					Homepage:    "TBD",
				},
				Extras: map[string]interface{}{
					"source": map[string]interface{}{
						"tools":    []interface{}{"SmartPy"},
						"location": "TBD",
					},
					"permissions": map[string]interface{}{
						"operator": "owner-or-operator-transfer",
						"receiver": "owner-no-hook",
						"sender":   "owner-no-hook",
					},
					"fa2-smartpy": map[string]interface{}{
						"configuration": map[string]interface{}{
							"add_mutez_transfer":           false,
							"allow_self_transfer":          false,
							"assume_consecutive_token_ids": true,
							"force_layouts":                true,
							"lazy_entry_points":            false,
							"lazy_entry_points_multiple":   false,
							"name":                         "FA2",
							"non_fungible":                 false,
							"readable":                     true,
							"single_asset":                 false,
							"store_total_supply":           true,
							"support_operator":             true,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bufTzip)
			if err := buf.UnmarshalJSON([]byte(tt.data)); (err != nil) != tt.wantErr {
				t.Errorf("bufTzip.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, buf, tt.want)
		})
	}
}
