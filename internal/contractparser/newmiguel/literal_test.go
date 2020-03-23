package newmiguel

import (
	"encoding/json"
	"testing"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

func TestLiteralContract(t *testing.T) {
	testCases := []struct {
		name    string
		rawJSON string
		path    string
		rawMeta string
		isRoot  bool

		expPrim string
		expType string
		expVal  string
	}{
		{
			name:    "contract/KT",
			rawJSON: `{"bytes": "016f516588d2ee560385e386708a13bd63da907cf300"}`,
			path:    "0/0/1/0/1",
			rawMeta: `{"0/0/1/0/1":{"prim":"contract","parameter":"{\"prim\":\"nat\"}","type":"contract"}}`,
			isRoot:  false,
			expPrim: consts.CONTRACT,
			expType: consts.CONTRACT,
			expVal:  "KT1JjN5bTE9yayzYHiBm6ruktwEWSHRF8aDm",
		},
		{
			name:    "contract/tz3",
			rawJSON: `{"bytes": "0002358cbffa97149631cfb999fa47f0035fb1ea8636"}`,
			path:    "0/1/1/o/0",
			rawMeta: `{"0/1/1/o/0":{"fieldname":"pour_dest","prim":"contract","parameter":"{\"prim\":\"nat\"}","type":"contract","name":"pour_dest"}}`,
			isRoot:  false,
			expPrim: consts.CONTRACT,
			expType: consts.CONTRACT,
			expVal:  "tz3RDC3Jdn4j15J7bBHZd29EUee9gVB1CxD9",
		},
	}

	decoder := new(literalDecoder)

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			var metadata meta.Metadata
			if err := json.Unmarshal([]byte(tt.rawMeta), &metadata); err != nil {
				t.Errorf("Invalid metadata string: %v %s", err, tt.rawMeta)
				return
			}
			nodeMetadata := metadata[tt.path]
			jsonData := gjson.Parse(tt.rawJSON)

			node, err := decoder.Decode(jsonData, tt.path, nodeMetadata, metadata, tt.isRoot)
			if err != nil {
				t.Errorf("Decode error: %v", err)
				return
			}

			if node.Prim != tt.expPrim {
				t.Errorf("Invalid prim. Got: %v, Expected: %v", node.Prim, tt.expPrim)
			}

			if node.Type != tt.expType {
				t.Errorf("Invalid type. Got: %v, Expected: %v", node.Type, tt.expType)
			}

			if res, ok := node.Value.(string); !ok || res != tt.expVal {
				t.Errorf("Invalid value. Got: %v, Expected: %v", res, tt.expVal)
			}
		})
	}
}

func TestLiteralChainID(t *testing.T) {
	testCases := []struct {
		name    string
		rawJSON string
		path    string
		rawMeta string
		isRoot  bool

		expPrim string
		expType string
		expVal  string
	}{
		{
			name:    "chainID/main",
			rawJSON: `{"bytes": "7a06a770"}`,
			path:    "0/o",
			rawMeta: `{"0":{"fieldname":"root","prim":"option","type":"option"},"0/o":{"prim":"chain_id","type":"chain_id"}}`,
			isRoot:  false,
			expPrim: consts.CHAINID,
			expType: consts.CHAINID,
			expVal:  "NetXdQprcVkpaWU",
		},
		{
			name:    "chainID/carthage",
			rawJSON: `{"bytes": "9caecab9"}`,
			path:    "0/o",
			rawMeta: `{"0":{"fieldname":"root","prim":"option","type":"option"},"0/o":{"prim":"chain_id","type":"chain_id"}}`,
			isRoot:  false,
			expPrim: consts.CHAINID,
			expType: consts.CHAINID,
			expVal:  "NetXjD3HPJJjmcd",
		},
	}

	decoder := new(literalDecoder)

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			var metadata meta.Metadata
			if err := json.Unmarshal([]byte(tt.rawMeta), &metadata); err != nil {
				t.Errorf("Invalid metadata string: %v %s", err, tt.rawMeta)
				return
			}
			nodeMetadata := metadata[tt.path]
			jsonData := gjson.Parse(tt.rawJSON)

			node, err := decoder.Decode(jsonData, tt.path, nodeMetadata, metadata, tt.isRoot)
			if err != nil {
				t.Errorf("Decode error: %v", err)
				return
			}

			if node.Prim != tt.expPrim {
				t.Errorf("Invalid prim. Got: %v, Expected: %v", node.Prim, tt.expPrim)
			}

			if node.Type != tt.expType {
				t.Errorf("Invalid type. Got: %v, Expected: %v", node.Type, tt.expType)
			}

			if res, ok := node.Value.(string); !ok || res != tt.expVal {
				t.Errorf("Invalid value. Got: %v, Expected: %v", res, tt.expVal)
			}
		})
	}
}

func TestLiteralSignature(t *testing.T) {
	testCases := []struct {
		name    string
		rawJSON string
		path    string
		rawMeta string
		isRoot  bool

		expPrim string
		expType string
		expVal  string
	}{
		{
			name:    "signature",
			rawJSON: `{"bytes": "bdc36db614aaa6084549020d376bb2469b5ea888dca2f7afbe5a0095bcc45ca0d8b5f00a051969437fe092debbcfe19d66378fbb74104de7eb1ecd895a64a80a"}`,
			path:    "0/1/1/l/o",
			rawMeta: `{"0":{"prim":"or","args":["0/0","0/1"],"type":"namedunion"},"0/0":{"fieldname":"default","prim":"unit","type":"unit","name":"default"},"0/1":{"fieldname":"main","prim":"pair","args":["0/1/0","0/1/1"],"type":"namedtuple","name":"main"},"0/1/0":{"typename":"payload","prim":"pair","args":["0/1/0/0","0/1/0/1"],"type":"namedtuple","name":"payload"},"0/1/0/0":{"fieldname":"counter","prim":"nat","type":"nat","name":"counter"},"0/1/0/1":{"typename":"action","prim":"or","args":["0/1/0/1/0","0/1/0/1/1"],"type":"namedunion","name":"action"},"0/1/0/1/0":{"fieldname":"operation","prim":"lambda","parameter":"{\"prim\":\"unit\"}","type":"lambda","name":"operation"},"0/1/0/1/1":{"fieldname":"change_keys","prim":"pair","args":["0/1/0/1/1/0","0/1/0/1/1/1"],"type":"namedtuple","name":"change_keys"},"0/1/0/1/1/0":{"fieldname":"threshold","prim":"nat","type":"nat","name":"threshold"},"0/1/0/1/1/1":{"fieldname":"keys","prim":"list","type":"list","name":"keys"},"0/1/0/1/1/1/l":{"prim":"key","type":"key"},"0/1/1":{"fieldname":"sigs","prim":"list","type":"list","name":"sigs"},"0/1/1/l":{"prim":"option","type":"option"},"0/1/1/l/o":{"prim":"signature","type":"signature"}}`,
			isRoot:  false,
			expPrim: consts.SIGNATURE,
			expType: consts.SIGNATURE,
			expVal:  "signpEFVQ1rW3TnVhc3PXf6SHRj7PvxwfJhBukWfB5X9rDhzpEk3ms5gRh763e922n52uQcjeqhqPdYi7WbFs2ERrNAPmCZJ",
		},
	}

	decoder := new(literalDecoder)

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			var metadata meta.Metadata
			if err := json.Unmarshal([]byte(tt.rawMeta), &metadata); err != nil {
				t.Errorf("Invalid metadata string: %v %s", err, tt.rawMeta)
				return
			}
			nodeMetadata := metadata[tt.path]
			jsonData := gjson.Parse(tt.rawJSON)

			node, err := decoder.Decode(jsonData, tt.path, nodeMetadata, metadata, tt.isRoot)
			if err != nil {
				t.Errorf("Decode error: %v", err)
				return
			}

			if node.Prim != tt.expPrim {
				t.Errorf("Invalid prim. Got: %v, Expected: %v", node.Prim, tt.expPrim)
			}

			if node.Type != tt.expType {
				t.Errorf("Invalid type. Got: %v, Expected: %v", node.Type, tt.expType)
			}

			if res, ok := node.Value.(string); !ok || res != tt.expVal {
				t.Errorf("Invalid value. Got: %v, Expected: %v", res, tt.expVal)
			}
		})
	}
}
