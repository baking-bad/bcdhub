package protocols

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/parsers/contract"
	"github.com/baking-bad/bcdhub/internal/parsers/migrations"
	"github.com/baking-bad/bcdhub/internal/parsers/storage"
	"github.com/pkg/errors"
)

// Specific -
type Specific struct {
	StorageParser         storage.Parser
	ContractParser        contract.Parser
	MigrationParser       migrations.Parser
	NeedReceiveRawStorage bool
}

// Get -
func Get(ctx *config.Context, protocol string) (*Specific, error) {
	switch protocol {
	case "ProtoGenesisGenesisGenesisGenesisGenesisGenesk612im",
		"PrihK96nBAFSxVL1GLJTVhu9YnzkMFiBeuJRPA8NwuZVZCE1L6i",
		"PtBMwNZT94N7gXKw4i273CKcSaBrrBnqnt3RATExNKr9KNX2USV",
		"ProtoDemoNoopsDemoNoopsDemoNoopsDemoNoopsDemo6XBoYp",
		"PtYuensgYBb3G3x1hLLbCmcav8ue8Kyd2khADcL5LsT5R1hcXex",
		"Ps9mPmXaRzmzk35gbAYNCAw6UXdE2qoABTHbN2oEEc1qM7CwT9P",
		"PsYLVpVvgbLhAhoqAkMFUo6gudkJ9weNXhUYCiLDzcUpFpkk8Wt",
		"PsddFKi32cMJ2qPjf43Qv5GDWLDPZb3T3bF6fLKiF5HtvHNU7aP",
		"Pt24m4xiPbLDhVgVfABUjirbmda3yohdN82Sp9FeuAXJ4eV9otd",
		"PtCJ7pwoxe8JasnHY8YonnLYjcVHmhiARPJvqcC6VfHT5s8k8sY":
		return &Specific{
			StorageParser:         storage.NewAlpha(),
			ContractParser:        contract.NewAlpha(ctx),
			MigrationParser:       migrations.NewAlpha(),
			NeedReceiveRawStorage: false,
		}, nil
	case "PsBabyM1eUXZseaJdmXFApDSBqj8YBfwELoxZHHW77EMcAbbwAS",
		"PsBABY5HQTSkA4297zNHfsZNKtxULfL18y95qb3m53QJiXGmrbU":
		return &Specific{
			StorageParser:         storage.NewBabylon(ctx.BigMapDiffs, ctx.Operations, ctx.Accounts),
			ContractParser:        contract.NewBabylon(ctx),
			MigrationParser:       migrations.NewBabylon(),
			NeedReceiveRawStorage: true,
		}, nil
	case "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
		"PryLyZ8A11FXDr1tRE9zQ7Di6Y8zX48RfFCFpkjC8Pt9yCBLhtN",
		"PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
		"PtEdoTezd3RHSC31mpxxo1npxFjoWWcFgQtxapi51Z8TLu6v6Uq",
		"PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA":
		return &Specific{
			StorageParser:         storage.NewBabylon(ctx.BigMapDiffs, ctx.Operations, ctx.Accounts),
			ContractParser:        contract.NewBabylon(ctx),
			MigrationParser:       migrations.NewCarthage(),
			NeedReceiveRawStorage: true,
		}, nil
	case "PrrUA9dCzbqBzugjQyw65HLHKjhH3HMFSLLHLZjj5rkmkG13Fej",
		"PsrsRVg1Gycjn5LvMtoYSQah1znvYmGp8bHLxwYLBZaYFf2CEkV",
		"PsFLorenaUUuikDWvMDr6fGBRG8kt3e3D3fHoXK1j1BFRxeSH4i",
		"PtGRANADsDU8R9daYKAgWnQYAJ64omN1o3KMGVCykShA97vQbvV":
		return &Specific{
			StorageParser:         storage.NewLazyBabylon(ctx.BigMapDiffs, ctx.Operations, ctx.Accounts),
			ContractParser:        contract.NewBabylon(ctx),
			MigrationParser:       migrations.NewCarthage(),
			NeedReceiveRawStorage: true,
		}, nil
	case "PtHangz2aRngywmSRGGvrcTyMbbdpWdpFKuS4uMWxg2RaH9i1qx",
		"PtHangzHogokSuiMHemCuowEavgYTP8J5qQ9fQS793MHYFpCY3r",
		"PsiThaCaT47Zboaw71QWScM8sXeMM7bbQFncK9FLqYc6EKdpjVP",
		"Psithaca2MLRFYargivpo7YvUr7wUDqyxrdhC5CQq78mRvimz6A":
		return &Specific{
			StorageParser:         storage.NewLazyBabylon(ctx.BigMapDiffs, ctx.Operations, ctx.Accounts),
			ContractParser:        contract.NewHangzhou(ctx),
			MigrationParser:       migrations.NewCarthage(),
			NeedReceiveRawStorage: true,
		}, nil
	case "PtJakart2xVj7pYXJBXrqHgd82rdkLey5ZeeGwDgPp9rhQUbSqY",
		"PtKathmankSpLLDALzWw7CGD2j2MtyveTwboEYokqUCP4a1LxMg",
		"ProtoALphaALphaALphaALphaALphaALphaALphaALphaDdp3zK",
		"PtLimaPtLMwfNinJi9rCfDPWea8dFgTZ1MeJ9f1m2SRic6ayiwW",
		"PtMumbaiiFFEGbew1rRjzSPyzRbA51Tm3RVZL5suHPxSZYDhCEc",
		"PtMumbai2TmsJHNGRkD8v8YDbtao7BLUC3wjASn1inAKLFCjaH1",
		"PtNairobiyssHuh87hEhfVBGCVrK3WnS8Z2FT4ymB5tAa4r1nQf",
		"ProxfordSW2S7fvchT1Zgj2avb5UES194neRyYVXoaDGvF9egt8",
		"ProxfordYmVfjWnRcgjWH36fW6PArwqykTFzotUxRs6gmTcZDuH",
		"PtParisA6ruu136piHaBC7cQLDP87JEqtczJWP2pLa5QCELGBH5",
		"PtParisBQscdCm6Cfow6ndeU6wKJyA3aV1j4D3gQBQMsTQyJCrz",
		"PtParisBxoLz5gzMmn3d9WBQNoPSZakgnkMC2VNuQ3KXfUtUQeZ",
		"PsParisCZo7KAh1Z1smVd9ZMZ1HHn5gkzbM94V3PLCpknFWhUAi",
		"PsQuebecnLByd3JwTiGadoG4nGWi3HYiLXUjkibeFV8dCFeVMUg",
		"PsRiotumaAMotcRoDWW1bysEhQy2n1M5fy8JgRp8jjRfHGmfeA7":
		return &Specific{
			StorageParser:         storage.NewLazyBabylon(ctx.BigMapDiffs, ctx.Operations, ctx.Accounts),
			ContractParser:        contract.NewJakarta(ctx),
			MigrationParser:       migrations.NewJakarta(),
			NeedReceiveRawStorage: true,
		}, nil

	default:
		return nil, errors.Errorf("unknown protocol in parser's creation: %s", protocol)

	}
}

// NeedImplicitParsing -
func NeedImplicitParsing(protocol string) bool {
	switch protocol {
	case "ProtoGenesisGenesisGenesisGenesisGenesisGenesk612im",
		"PrihK96nBAFSxVL1GLJTVhu9YnzkMFiBeuJRPA8NwuZVZCE1L6i",
		"PtBMwNZT94N7gXKw4i273CKcSaBrrBnqnt3RATExNKr9KNX2USV",
		"ProtoDemoNoopsDemoNoopsDemoNoopsDemoNoopsDemo6XBoYp",
		"PtYuensgYBb3G3x1hLLbCmcav8ue8Kyd2khADcL5LsT5R1hcXex",
		"Ps9mPmXaRzmzk35gbAYNCAw6UXdE2qoABTHbN2oEEc1qM7CwT9P",
		"PsYLVpVvgbLhAhoqAkMFUo6gudkJ9weNXhUYCiLDzcUpFpkk8Wt",
		"PsddFKi32cMJ2qPjf43Qv5GDWLDPZb3T3bF6fLKiF5HtvHNU7aP",
		"Pt24m4xiPbLDhVgVfABUjirbmda3yohdN82Sp9FeuAXJ4eV9otd",
		"PtCJ7pwoxe8JasnHY8YonnLYjcVHmhiARPJvqcC6VfHT5s8k8sY",
		"PsBabyM1eUXZseaJdmXFApDSBqj8YBfwELoxZHHW77EMcAbbwAS",
		"PsBABY5HQTSkA4297zNHfsZNKtxULfL18y95qb3m53QJiXGmrbU",
		"PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
		"PryLyZ8A11FXDr1tRE9zQ7Di6Y8zX48RfFCFpkjC8Pt9yCBLhtN",
		"PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
		"PtEdoTezd3RHSC31mpxxo1npxFjoWWcFgQtxapi51Z8TLu6v6Uq",
		"PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA",
		"PrrUA9dCzbqBzugjQyw65HLHKjhH3HMFSLLHLZjj5rkmkG13Fej",
		"PsrsRVg1Gycjn5LvMtoYSQah1znvYmGp8bHLxwYLBZaYFf2CEkV",
		"PsFLorenaUUuikDWvMDr6fGBRG8kt3e3D3fHoXK1j1BFRxeSH4i":
		return false
	case "PtGRANADsDU8R9daYKAgWnQYAJ64omN1o3KMGVCykShA97vQbvV",
		"PtHangz2aRngywmSRGGvrcTyMbbdpWdpFKuS4uMWxg2RaH9i1qx",
		"PtHangzHogokSuiMHemCuowEavgYTP8J5qQ9fQS793MHYFpCY3r",
		"PsiThaCaT47Zboaw71QWScM8sXeMM7bbQFncK9FLqYc6EKdpjVP",
		"Psithaca2MLRFYargivpo7YvUr7wUDqyxrdhC5CQq78mRvimz6A",
		"PtJakart2xVj7pYXJBXrqHgd82rdkLey5ZeeGwDgPp9rhQUbSqY",
		"PtKathmankSpLLDALzWw7CGD2j2MtyveTwboEYokqUCP4a1LxMg",
		"ProtoALphaALphaALphaALphaALphaALphaALphaALphaDdp3zK",
		"PtLimaPtLMwfNinJi9rCfDPWea8dFgTZ1MeJ9f1m2SRic6ayiwW",
		"PtMumbaiiFFEGbew1rRjzSPyzRbA51Tm3RVZL5suHPxSZYDhCEc",
		"PtMumbai2TmsJHNGRkD8v8YDbtao7BLUC3wjASn1inAKLFCjaH1",
		"PtNairobiyssHuh87hEhfVBGCVrK3WnS8Z2FT4ymB5tAa4r1nQf",
		"ProxfordSW2S7fvchT1Zgj2avb5UES194neRyYVXoaDGvF9egt8",
		"ProxfordYmVfjWnRcgjWH36fW6PArwqykTFzotUxRs6gmTcZDuH",
		"PtParisA6ruu136piHaBC7cQLDP87JEqtczJWP2pLa5QCELGBH5",
		"PtParisBQscdCm6Cfow6ndeU6wKJyA3aV1j4D3gQBQMsTQyJCrz",
		"PtParisBxoLz5gzMmn3d9WBQNoPSZakgnkMC2VNuQ3KXfUtUQeZ",
		"PsParisCZo7KAh1Z1smVd9ZMZ1HHn5gkzbM94V3PLCpknFWhUAi",
		"PsQuebecnLByd3JwTiGadoG4nGWi3HYiLXUjkibeFV8dCFeVMUg",
		"PsRiotumaAMotcRoDWW1bysEhQy2n1M5fy8JgRp8jjRfHGmfeA7":
		return true
	}
	return false
}
