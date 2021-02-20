package bcd

import (
	"github.com/pkg/errors"
)

// This is the list of protocols BCD supports
// Every time new protocol is proposed we determine if everything works fine or implement a custom handler otherwise
// After that we append protocol to this list with a corresponding handler id (aka symlink)
var symLinks = map[string]string{
	"ProtoGenesisGenesisGenesisGenesisGenesisGenesk612im": "alpha",
	"PrihK96nBAFSxVL1GLJTVhu9YnzkMFiBeuJRPA8NwuZVZCE1L6i": "alpha",
	"PtBMwNZT94N7gXKw4i273CKcSaBrrBnqnt3RATExNKr9KNX2USV": "alpha",
	"ProtoDemoNoopsDemoNoopsDemoNoopsDemoNoopsDemo6XBoYp": "alpha",
	"PtYuensgYBb3G3x1hLLbCmcav8ue8Kyd2khADcL5LsT5R1hcXex": "alpha",
	"Ps9mPmXaRzmzk35gbAYNCAw6UXdE2qoABTHbN2oEEc1qM7CwT9P": "alpha",
	"PsYLVpVvgbLhAhoqAkMFUo6gudkJ9weNXhUYCiLDzcUpFpkk8Wt": "alpha",
	"PsddFKi32cMJ2qPjf43Qv5GDWLDPZb3T3bF6fLKiF5HtvHNU7aP": "alpha",
	"Pt24m4xiPbLDhVgVfABUjirbmda3yohdN82Sp9FeuAXJ4eV9otd": "alpha",
	"PtCJ7pwoxe8JasnHY8YonnLYjcVHmhiARPJvqcC6VfHT5s8k8sY": "alpha",
	"PsBabyM1eUXZseaJdmXFApDSBqj8YBfwELoxZHHW77EMcAbbwAS": "babylon",
	"PsBABY5HQTSkA4297zNHfsZNKtxULfL18y95qb3m53QJiXGmrbU": "babylon",
	"PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb": "babylon", // Carthagenet
	"PryLyZ8A11FXDr1tRE9zQ7Di6Y8zX48RfFCFpkjC8Pt9yCBLhtN": "babylon", // Dalphanet
	"PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo": "babylon", // Delphinet
	"PtEdoTezd3RHSC31mpxxo1npxFjoWWcFgQtxapi51Z8TLu6v6Uq": "babylon", // Edonet 8.1
	"PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA": "babylon", // Edonet 8.2
}

// GetProtoSymLink -
func GetProtoSymLink(protocol string) (string, error) {
	if protoSymLink, ok := symLinks[protocol]; ok {
		return protoSymLink, nil
	}
	return "", errors.Errorf("Unknown protocol: %s", protocol)
}

// GetCurrentProtocol - returns last supported protocol
func GetCurrentProtocol() string {
	return "PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA"
}
