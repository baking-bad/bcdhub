package indexer

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestIsUpgradable(t *testing.T) {
	testCases := []struct {
		address string
		result  bool
	}{
		{
			address: "KT1CyJxNgctn3gQKBu9ivKN5RSgqpmEhX5W8",
			result:  true,
		},
		{
			address: "KT1G9SQK1YK8oDTJAWaPjuBmY2fX5QGBnYLj",
			result:  true,
		},
		{
			address: "KT18bwMJoY3xj6vdB94mLyGGasyNZmSgZBuT",
			result:  true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.address, func(t *testing.T) {
			paramPath := fmt.Sprintf("./testdata/metadata/%s/parameter.json", tt.address)
			storagePath := fmt.Sprintf("./testdata/metadata/%s/storage.json", tt.address)

			paramFile, err := ioutil.ReadFile(paramPath)
			if err != nil {
				t.Errorf("ioutil.ReadFile %v error %v", paramPath, err)
				return
			}

			storageFile, err := ioutil.ReadFile(storagePath)
			if err != nil {
				t.Errorf("ioutil.ReadFile %v error %v", storagePath, err)
				return
			}

			result, err := isUpgradable(string(storageFile), string(paramFile))
			if err != nil {
				t.Errorf("isUpgradable %v error: %v", tt.address, err)
				return
			}

			if result != tt.result {
				t.Errorf("invalid result %v", tt.address)
			}
		})
	}

}
