package fetch

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

const (
	contractFormatPath = "%s/contracts/%s/%s_%s.json"
)

// ContractWithRPC -
func ContractWithRPC(rpc noderpc.INode, address, network, protocol, filesDirectory string, fallbackLevel int64) (gjson.Result, error) {
	if filesDirectory != "" {
		data, err := Contract(address, network, protocol, filesDirectory)
		switch {
		case err == nil:
			return gjson.ParseBytes(data), nil
		case !os.IsNotExist(err):
			return gjson.Result{}, err
		}
	}
	return rpc.GetScriptJSON(address, fallbackLevel)
}

// RemoveContractFromFileSystem -
func RemoveContractFromFileSystem(address, network, protocol, filesDirectory string) error {
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
		return nil, err
	}
	return ioutil.ReadFile(filePath)
}
