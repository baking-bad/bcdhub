package bcd

import (
	"github.com/pkg/errors"
)

// This is the list of protocols BCD supports
// Every time new protocol is proposed we determine if everything works fine or implement a custom handler otherwise
// After that we append protocol to this list with a corresponding handler id (aka symlink)
var symLinks = map[string]string{
	"ProtoGenesisGenesisGenesisGenesisGenesisGenesk612im": SymLinkAlpha,
	"PrihK96nBAFSxVL1GLJTVhu9YnzkMFiBeuJRPA8NwuZVZCE1L6i": SymLinkAlpha,
	"PtBMwNZT94N7gXKw4i273CKcSaBrrBnqnt3RATExNKr9KNX2USV": SymLinkAlpha,
	"ProtoDemoNoopsDemoNoopsDemoNoopsDemoNoopsDemo6XBoYp": SymLinkAlpha,
	"PtYuensgYBb3G3x1hLLbCmcav8ue8Kyd2khADcL5LsT5R1hcXex": SymLinkAlpha,
	"Ps9mPmXaRzmzk35gbAYNCAw6UXdE2qoABTHbN2oEEc1qM7CwT9P": SymLinkAlpha,
	"PsYLVpVvgbLhAhoqAkMFUo6gudkJ9weNXhUYCiLDzcUpFpkk8Wt": SymLinkAlpha,
	"PsddFKi32cMJ2qPjf43Qv5GDWLDPZb3T3bF6fLKiF5HtvHNU7aP": SymLinkAlpha,
	"Pt24m4xiPbLDhVgVfABUjirbmda3yohdN82Sp9FeuAXJ4eV9otd": SymLinkAlpha,
	"PtCJ7pwoxe8JasnHY8YonnLYjcVHmhiARPJvqcC6VfHT5s8k8sY": SymLinkAlpha,
	"PsBabyM1eUXZseaJdmXFApDSBqj8YBfwELoxZHHW77EMcAbbwAS": SymLinkBabylon,
	"PsBABY5HQTSkA4297zNHfsZNKtxULfL18y95qb3m53QJiXGmrbU": SymLinkBabylon,
	"PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb": SymLinkBabylon, // Carthagenet
	"PryLyZ8A11FXDr1tRE9zQ7Di6Y8zX48RfFCFpkjC8Pt9yCBLhtN": SymLinkBabylon, // Dalphanet
	"PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo": SymLinkBabylon, // Delphinet
	"PtEdoTezd3RHSC31mpxxo1npxFjoWWcFgQtxapi51Z8TLu6v6Uq": SymLinkBabylon, // Edonet 8.1
	"PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA": SymLinkBabylon, // Edonet 8.2
	"PrrUA9dCzbqBzugjQyw65HLHKjhH3HMFSLLHLZjj5rkmkG13Fej": SymLinkBabylon, // Falphanet
	"PsrsRVg1Gycjn5LvMtoYSQah1znvYmGp8bHLxwYLBZaYFf2CEkV": SymLinkBabylon, // Falphanet
	"PsFLorenaUUuikDWvMDr6fGBRG8kt3e3D3fHoXK1j1BFRxeSH4i": SymLinkBabylon, // Florencenet (no baking accounts)
	"PtGRANADsDU8R9daYKAgWnQYAJ64omN1o3KMGVCykShA97vQbvV": SymLinkBabylon, // Granadanet
	"PtHangzHogokSuiMHemCuowEavgYTP8J5qQ9fQS793MHYFpCY3r": SymLinkBabylon, // Hangzhounet
	"PtHangz2aRngywmSRGGvrcTyMbbdpWdpFKuS4uMWxg2RaH9i1qx": SymLinkBabylon, // Hangzhounet 2
	"PsiThaCaT47Zboaw71QWScM8sXeMM7bbQFncK9FLqYc6EKdpjVP": SymLinkBabylon, // Itacanet
	"Psithaca2MLRFYargivpo7YvUr7wUDqyxrdhC5CQq78mRvimz6A": SymLinkBabylon, // Itacanet 2
	"ProtoALphaALphaALphaALphaALphaALphaALphaALphaDdp3zK": SymLinkJakarta, // Weekly and daily
	"PtJakart2xVj7pYXJBXrqHgd82rdkLey5ZeeGwDgPp9rhQUbSqY": SymLinkJakarta, // Jakarta
	"PtKathmankSpLLDALzWw7CGD2j2MtyveTwboEYokqUCP4a1LxMg": SymLinkJakarta, // Kathmandu
	"PtLimaPtLMwfNinJi9rCfDPWea8dFgTZ1MeJ9f1m2SRic6ayiwW": SymLinkJakarta, // Lima
	"PtMumbaiiFFEGbew1rRjzSPyzRbA51Tm3RVZL5suHPxSZYDhCEc": SymLinkJakarta, // Mumbai
	"PtMumbai2TmsJHNGRkD8v8YDbtao7BLUC3wjASn1inAKLFCjaH1": SymLinkJakarta, // Mumbai 2
	"PtNairobiyssHuh87hEhfVBGCVrK3WnS8Z2FT4ymB5tAa4r1nQf": SymLinkJakarta, // Nairobinet
	"ProxfordSW2S7fvchT1Zgj2avb5UES194neRyYVXoaDGvF9egt8": SymLinkJakarta, // Oxford
	"ProxfordYmVfjWnRcgjWH36fW6PArwqykTFzotUxRs6gmTcZDuH": SymLinkJakarta, // Oxford 2
	"PtParisA6ruu136piHaBC7cQLDP87JEqtczJWP2pLa5QCELGBH5": SymLinkJakarta, // Paris A
	"PtParisBQscdCm6Cfow6ndeU6wKJyA3aV1j4D3gQBQMsTQyJCrz": SymLinkJakarta, // Paris B
	"PtParisBxoLz5gzMmn3d9WBQNoPSZakgnkMC2VNuQ3KXfUtUQeZ": SymLinkJakarta, // Paris B 2,
	"PsParisCZo7KAh1Z1smVd9ZMZ1HHn5gkzbM94V3PLCpknFWhUAi": SymLinkJakarta, // Paris C
	"PsQuebecnLByd3JwTiGadoG4nGWi3HYiLXUjkibeFV8dCFeVMUg": SymLinkJakarta, // Quebec
	"PsRiotumaAMotcRoDWW1bysEhQy2n1M5fy8JgRp8jjRfHGmfeA7": SymLinkJakarta, // Rionet
	"PtSeouLouXkxhg39oWzjxDWaCydNfR3RxCUrNe4Q9Ro8BTehcbh": SymLinkJakarta, // Seoulnet
	"PtTALLiNtPec7mE7yY4m3k26J8Qukef3E3ehzhfXgFZKGtDdAXu": SymLinkJakarta, // Tallinnnet
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
	return "PtSeouLouXkxhg39oWzjxDWaCydNfR3RxCUrNe4Q9Ro8BTehcbh"
}

// SymLink - returns last sym link
func SymLink() (string, error) {
	return GetProtoSymLink(GetCurrentProtocol())
}

// Symbolic links
const (
	SymLinkAlpha   = "alpha"
	SymLinkBabylon = "babylon"
	SymLinkJakarta = "jakarta"
)
