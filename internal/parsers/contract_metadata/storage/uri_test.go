package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSha256URI_Parse(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		hash    string
		link    string
		wantErr bool
	}{
		{
			name:  "test 1",
			value: "sha256://0xeaa42ea06b95d7917d22135a630e65352cfd0a721ae88155a1512468a95cb750/https:%2F%2Ftezos.com",
			hash:  "0xeaa42ea06b95d7917d22135a630e65352cfd0a721ae88155a1512468a95cb750",
			link:  "https://tezos.com",
		}, {
			name:  "test 2",
			value: "sha256://0xeaa42ea06b95d7917d22135a630e65352cfd0a721ae88155a1512468a95cb750/https://tezos.com",
			hash:  "0xeaa42ea06b95d7917d22135a630e65352cfd0a721ae88155a1512468a95cb750",
			link:  "https://tezos.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri := &Sha256URI{}
			if err := uri.Parse(tt.value); (err != nil) != tt.wantErr {
				t.Errorf("Sha256URI.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, tt.hash, uri.Hash) {
				t.Errorf("Sha256URI.Parse() hash = %v, want %v", uri.Hash, tt.hash)
				return
			}
			if !assert.Equal(t, tt.link, uri.Link) {
				t.Errorf("Sha256URI.Parse() link = %v, want %v", uri.Hash, tt.hash)
				return
			}
		})
	}
}

func TestTezosStorageURI_Parse(t *testing.T) {
	type fields struct {
		Address string
		Network string
		Key     string
	}
	tests := []struct {
		name    string
		fields  fields
		value   string
		wantErr bool
	}{
		{
			name:  "test 1",
			value: "tezos-storage:hello",
			fields: fields{
				Address: "",
				Network: "",
				Key:     "hello",
			},
		}, {
			name:  "test 2",
			value: "tezos-storage://KT1QDFEu8JijYbsJqzoXq7mKvfaQQamHD1kX/foo",
			fields: fields{
				Address: "KT1QDFEu8JijYbsJqzoXq7mKvfaQQamHD1kX",
				Network: "",
				Key:     "foo",
			},
		}, {
			name:  "test 3",
			value: "tezos-storage://KT1QDFEu8JijYbsJqzoXq7mKvfaQQamHD1kX/%2Ffoo",
			fields: fields{
				Address: "KT1QDFEu8JijYbsJqzoXq7mKvfaQQamHD1kX",
				Network: "",
				Key:     "/foo",
			},
		}, {
			name:  "test 4",
			value: "tezos-storage://KT1QDFEu8JijYbsJqzoXq7mKvfaQQamHD1kX.mainnet/%2Ffoo",
			fields: fields{
				Address: "KT1QDFEu8JijYbsJqzoXq7mKvfaQQamHD1kX",
				Network: "mainnet",
				Key:     "/foo",
			},
		}, {
			name:  "test 5",
			value: "tezos-storage:metadata",
			fields: fields{
				Address: "",
				Network: "",
				Key:     "metadata",
			},
		}, {
			name:  "test 6",
			value: "tezos-storage:token_1_metadata",
			fields: fields{
				Address: "",
				Network: "",
				Key:     "token_1_metadata",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri := &TezosStorageURI{}
			if err := uri.Parse(tt.value); (err != nil) != tt.wantErr {
				t.Errorf("TezosStorageURI.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, tt.fields.Address, uri.Address) {
				t.Errorf("Sha256URI.Parse() address = %v, want %v", uri.Address, tt.fields.Address)
				return
			}
			if !assert.Equal(t, tt.fields.Network, uri.Network) {
				t.Errorf("Sha256URI.Parse() network = %v, want %v", uri.Network, tt.fields.Network)
				return
			}
			if !assert.Equal(t, tt.fields.Key, uri.Key) {
				t.Errorf("Sha256URI.Parse() key = %v, want %v", uri.Key, tt.fields.Key)
				return
			}
		})
	}
}
