package fetch

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/pkg/errors"
)

const (
	contractFormatPath    = "%s/contracts/%s/%s_%s.json"
	delegatorContractPath = "%s/contracts/scripts/b5d01d3bf75d0cc4b88e5f881074084c5e76a0b1f8bdab6249c591a6f45a314283d32fe01c7300563e82866ef1f0c34d401b9e00cba6d2018fe56595f06b5f02.json"
)

// RemoveContract -
func RemoveContract(address, network, protocol, filesDirectory string) error {
	if filesDirectory == "" {
		return errors.Errorf("Invalid filesDirectory: %s", filesDirectory)
	}
	protoSymLink, err := bcd.GetProtoSymLink(protocol)
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf(contractFormatPath, filesDirectory, network, address, protoSymLink)
	if _, err = os.Stat(filePath); err == nil {
		return os.Remove(filePath)
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

// RemoveAllContracts -
func RemoveAllContracts(network, filesDirectory string) error {
	if filesDirectory == "" {
		return errors.Errorf("Invalid filesDirectory: %s", filesDirectory)
	}

	dirPath := fmt.Sprintf("%s/contracts/%s", filesDirectory, network)
	if _, err := os.Stat(dirPath); err == nil {
		return os.RemoveAll(dirPath)
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Contract - reads contract from file system
func Contract(address, network, protocol, filesDirectory string) ([]byte, error) {
	if protocol == "" {
		protocol = bcd.GetCurrentProtocol()
	}
	protoSymLink, err := bcd.GetProtoSymLink(protocol)
	if err != nil {
		return nil, err
	}

	filePath := fmt.Sprintf(contractFormatPath, filesDirectory, network, address, protoSymLink)
	if _, err = os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			filePath = fmt.Sprintf(delegatorContractPath, filesDirectory)
		} else {
			return nil, err
		}
	}
	return ioutil.ReadFile(filePath)
}
