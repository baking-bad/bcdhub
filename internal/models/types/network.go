package types

import (
	"database/sql/driver"
	"strconv"

	"github.com/pkg/errors"
)

// Network -
type Network int64

// Network names
const (
	Empty Network = iota
	Mainnet
	Carthagenet
	Delphinet
	Edo2net
	Florencenet
	Granadanet
	Sandboxnet
	Hangzhounet
)

var networkNames = map[Network]string{
	Mainnet:     "mainnet",
	Carthagenet: "carthagenet",
	Delphinet:   "delphinet",
	Edo2net:     "edo2net",
	Florencenet: "florencenet",
	Granadanet:  "granadanet",
	Sandboxnet:  "sandboxnet",
	Hangzhounet: "hangzhounet",
}

var namesToNetwork = map[string]Network{
	"mainnet":     Mainnet,
	"carthagenet": Carthagenet,
	"delphinet":   Delphinet,
	"edo2net":     Edo2net,
	"florencenet": Florencenet,
	"granadanet":  Granadanet,
	"sandboxnet":  Sandboxnet,
	"hangzhounet": Hangzhounet,
}

// String - convert enum to string for printing
func (network Network) String() string {
	return networkNames[network]
}

// Scan -
func (network *Network) Scan(value interface{}) error {
	*network = Network(value.(int64))
	return nil
}

// Value -
func (network Network) Value() (driver.Value, error) { return uint64(network), nil }

// UnmarshalJSON -
func (network *Network) UnmarshalJSON(data []byte) error {
	name, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}
	newValue, ok := namesToNetwork[name]
	if !ok {
		return errors.Errorf("Unknown network: %d", network)
	}

	*network = newValue
	return nil
}

// MarshalJSON -
func (network Network) MarshalJSON() ([]byte, error) {
	name, ok := networkNames[network]
	if !ok {
		return nil, errors.Errorf("Unknown network: %d", network)
	}

	return []byte(strconv.Quote(name)), nil
}

// NewNetwork -
func NewNetwork(name string) Network {
	return namesToNetwork[name]
}
