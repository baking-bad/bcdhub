package types

import (
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
	Hangzhou2net
	Ithacanet
	Jakartanet
	Ghostnet
	Kathmandunet
	Limanet
	Mondaynet
	Dailynet
	Mumbainet
	Rollupnet
	Nairobinet
	Oxfordnet
	Weeklynet
	Parisnet
	Quebecnet
	Rionet
	Seoulnet
	Shadownet
	Tallinnnet
	Tezlink
)

var networkNames = map[Network]string{
	Mainnet:      "mainnet",
	Carthagenet:  "carthagenet",
	Delphinet:    "delphinet",
	Edo2net:      "edo2net",
	Florencenet:  "florencenet",
	Granadanet:   "granadanet",
	Sandboxnet:   "sandboxnet",
	Hangzhounet:  "hangzhounet",
	Hangzhou2net: "hangzhou2net",
	Ithacanet:    "ithacanet",
	Jakartanet:   "jakartanet",
	Ghostnet:     "ghostnet",
	Kathmandunet: "kathmandunet",
	Limanet:      "limanet",
	Mondaynet:    "mondaynet",
	Weeklynet:    "weeklynet",
	Dailynet:     "dailynet",
	Mumbainet:    "mumbainet",
	Rollupnet:    "rollupnet",
	Nairobinet:   "nairobinet",
	Oxfordnet:    "oxfordnet",
	Parisnet:     "parisnet",
	Quebecnet:    "quebecnet",
	Rionet:       "rionet",
	Seoulnet:     "seoulnet",
	Shadownet:    "shadownet",
	Tallinnnet:   "tallinnnet",
	Tezlink:      "tezlink",
}

var namesToNetwork = map[string]Network{
	"mainnet":      Mainnet,
	"carthagenet":  Carthagenet,
	"delphinet":    Delphinet,
	"edo2net":      Edo2net,
	"florencenet":  Florencenet,
	"granadanet":   Granadanet,
	"sandboxnet":   Sandboxnet,
	"hangzhounet":  Hangzhounet,
	"hangzhou2net": Hangzhou2net,
	"ithacanet":    Ithacanet,
	"jakartanet":   Jakartanet,
	"ghostnet":     Ghostnet,
	"kathmandunet": Kathmandunet,
	"limanet":      Limanet,
	"mondaynet":    Mondaynet,
	"dailynet":     Dailynet,
	"mumbainet":    Mumbainet,
	"rollupnet":    Rollupnet,
	"nairobinet":   Nairobinet,
	"oxfordnet":    Oxfordnet,
	"weeklynet":    Weeklynet,
	"parisnet":     Parisnet,
	"quebecnet":    Quebecnet,
	"rionet":       Rionet,
	"seoulnet":     Seoulnet,
	"shadownet":    Shadownet,
	"tallinnnet":   Tallinnnet,
	"tezlink":      Tezlink,
}

// String - convert enum to string for printing
func (network Network) String() string {
	return networkNames[network]
}

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

// Networks -
type Networks []Network

func (n Networks) Len() int           { return len(n) }
func (n Networks) Less(i, j int) bool { return n[i] < n[j] }
func (n Networks) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
