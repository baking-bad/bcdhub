package contractparser

import (
	"io/ioutil"
	"testing"

	"github.com/tidwall/gjson"
)

func TestTagFA12(t *testing.T) {
	fa12tag := "fa12"

	testCases := []struct {
		name string
		path string
		res  bool
	}{
		{
			name: "babylonnet/KT1Q3XGrpbqhF6ny4qLhuiKekmk86hiAnmhh",
			path: "testdata/tags/fa12-babylonnet-KT1Q3XGrpbqhF6ny4qLhuiKekmk86hiAnmhh.json",
			res:  true,
		},
		{
			name: "babylonnet/KT1XNvgLWxTpEZ88ruN16uxtRktFtAFyDnA1",
			path: "testdata/tags/fa12-babylonnet-KT1XNvgLWxTpEZ88ruN16uxtRktFtAFyDnA1.json",
			res:  true,
		},
		{
			name: "carthagenet/KT1VZsNB1nNmqMgAzHVMrJgHp28ozWnmSStB",
			path: "testdata/tags/fa12-carthagenet-KT1VZsNB1nNmqMgAzHVMrJgHp28ozWnmSStB.json",
			res:  true,
		},
		{
			name: "mainnet/KT1VG2WtYdSWz5E7chTeAdDPZNy2MpP8pTfL [wrong]",
			path: "testdata/tags/fa12-mainnet-KT1VG2WtYdSWz5E7chTeAdDPZNy2MpP8pTfL.json",
			res:  false,
		},
		{
			name: "babylonnet/KT1C5EUxReJJKssL9x9NvhkNEnMQnx9kr85D [wrong]",
			path: "testdata/tags/fa12-babylonnet-KT1C5EUxReJJKssL9x9NvhkNEnMQnx9kr85D.json",
			res:  false,
		},
		{
			name: "carthagenet/KT19DKAoPmtdoqg3WJKmKRw3r1Ki7u4U6tzm [wrong]",
			path: "testdata/tags/fa12-carthagenet-KT19DKAoPmtdoqg3WJKmKRw3r1Ki7u4U6tzm.json",
			res:  false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			file, err := ioutil.ReadFile(tt.path)
			if err != nil {
				t.Errorf("ioutil.ReadFile %v error %v", tt.path, err)
				return
			}

			parsed := gjson.ParseBytes(file)
			p, err := newParameter(parsed)
			if err != nil {
				t.Errorf("newParameter error %v", err)
				return
			}

			if _, ok := p.Tags[fa12tag]; tt.res != ok {
				t.Errorf("Wrong res. Got: %v", p.Tags)
			}
		})
	}

}
