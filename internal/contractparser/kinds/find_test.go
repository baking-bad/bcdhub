package kinds

import (
	"io/ioutil"
	"testing"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

func testFile(t *testing.T, tag, path string, res bool) error {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Errorf("ioutil.ReadFile %v error %v", path, err)
	}

	parsed := gjson.ParseBytes(file)
	m, err := meta.ParseMetadata(parsed)
	if err != nil {
		return errors.Errorf("meta.ParseMetadata %v error %v", path, err)
	}

	interfaces, err := Load()
	if err != nil {
		return errors.Errorf("newParameter error %v", err)
	}

	tags, err := Find(m, interfaces)
	if err != nil {
		return errors.Errorf("Find error %v", err)
	}

	if ok := helpers.StringInArray(tag, tags); res != ok {
		return errors.Errorf("ok != res: %v != %v", ok, res)
	}
	return nil
}

func TestTagFA12(t *testing.T) {
	testCases := []struct {
		name string
		path string
		res  bool
	}{
		{
			name: "babylonnet/KT1Q3XGrpbqhF6ny4qLhuiKekmk86hiAnmhh",
			path: "testdata/fa12-babylonnet-KT1Q3XGrpbqhF6ny4qLhuiKekmk86hiAnmhh.json",
			res:  true,
		},
		{
			name: "babylonnet/KT1XNvgLWxTpEZ88ruN16uxtRktFtAFyDnA1",
			path: "testdata/fa12-babylonnet-KT1XNvgLWxTpEZ88ruN16uxtRktFtAFyDnA1.json",
			res:  true,
		},
		{
			name: "carthagenet/KT1VZsNB1nNmqMgAzHVMrJgHp28ozWnmSStB",
			path: "testdata/fa12-carthagenet-KT1VZsNB1nNmqMgAzHVMrJgHp28ozWnmSStB.json",
			res:  true,
		},
		{
			name: "mainnet/KT1VG2WtYdSWz5E7chTeAdDPZNy2MpP8pTfL [wrong]",
			path: "testdata/fa12-mainnet-KT1VG2WtYdSWz5E7chTeAdDPZNy2MpP8pTfL.json",
			res:  false,
		},
		{
			name: "babylonnet/KT1C5EUxReJJKssL9x9NvhkNEnMQnx9kr85D [wrong]",
			path: "testdata/fa12-babylonnet-KT1C5EUxReJJKssL9x9NvhkNEnMQnx9kr85D.json",
			res:  false,
		},
		{
			name: "carthagenet/KT19DKAoPmtdoqg3WJKmKRw3r1Ki7u4U6tzm [wrong]",
			path: "testdata/fa12-carthagenet-KT19DKAoPmtdoqg3WJKmKRw3r1Ki7u4U6tzm.json",
			res:  false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if err := testFile(t, "fa12", tt.path, tt.res); err != nil {
				t.Error(err)
			}
		})
	}

}

func TestTagFA2(t *testing.T) {
	testCases := []struct {
		name string
		path string
		res  bool
	}{
		{
			name: "carthagenet/KT19nsmdVr54y2MLG1zKtbRUc6TqLfdYXNRG",
			path: "testdata/fa2-carthagenet-KT19nsmdVr54y2MLG1zKtbRUc6TqLfdYXNRG.json",
			res:  true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if err := testFile(t, "fa2", tt.path, tt.res); err != nil {
				t.Error(err)
			}
		})
	}

}

func TestTagViewNat(t *testing.T) {
	testCases := []struct {
		name string
		path string
		res  bool
	}{
		{
			name: "mainnet/KT1NhtHwHD5cqabfSdwg1Fowud5f175eShwx",
			path: "testdata/viewnat-mainnet-KT1NhtHwHD5cqabfSdwg1Fowud5f175eShwx.json",
			res:  true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if err := testFile(t, "view_nat", tt.path, tt.res); err != nil {
				t.Error(err)
			}
		})
	}

}

func TestTagViewBalanceOf(t *testing.T) {
	testCases := []struct {
		name string
		path string
		res  bool
	}{
		{
			name: "carthagenet/KT1G8XaqpT934gPfwbgYBhULDpCmZw4VGwtc",
			path: "testdata/viewbalanceof-carthagenet-KT1G8XaqpT934gPfwbgYBhULDpCmZw4VGwtc.json",
			res:  true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if err := testFile(t, "view_balance_of", tt.path, tt.res); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestEmpty(t *testing.T) {
	testCases := []struct {
		name string
		path string
		res  bool
	}{
		{
			name: "mainnet/KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
			path: "testdata/empty-mainnet-KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM.json",
			res:  false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if err := testFile(t, "view_nat", tt.path, tt.res); err != nil {
				t.Error(err)
			}
		})
	}
}
