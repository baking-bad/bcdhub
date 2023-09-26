package migrations

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/contract"
	"github.com/baking-bad/bcdhub/internal/models"
	modelsContract "github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
)

// Jakarta -
type Jakarta struct {
	contracts map[string]struct{}
}

// NewJakarta -
func NewJakarta() *Jakarta {
	return &Jakarta{
		contracts: map[string]struct{}{
			"KT1MzfYSbq18fYr4f44aQRoZBQN72BAtiz5j": {},
			"KT1Kfbk3B6NYPCPohPBDU3Hxf5Xeyy9PdkNp": {},
			"KT1JW6PwhfaEJu6U3ENsxUeja48AdtqSoekd": {},
			"KT1VsSxSXUkgw6zkBGgUuDXXuJs9ToPqkrCg": {},
			"KT1TcAHw5gpejyemwRtdNyFKGBLc4qwA5gtw": {},
			"KT1FN5fcNNcgieGjzxbVEPWUpJGwZEpzNGA8": {},
			"KT1Um7ieBEytZtumecLqGeL56iY6BuWoBgio": {},
			"KT1QuofAgnsWffHzLA7D78rxytJruGHDe7XG": {},
			"KT1CSKPf2jeLpMmrgKquN2bCjBTkAcAdRVDy": {},
			"KT1D5NmtDtgCwPxYNb2ZK2But6dhNLs1T1bV": {},
			"KT1VvXEpeBpreAVpfp4V8ZujqWu2gVykwXBJ": {},
			"KT1TzamC1SCj68ia2E4q2GWZeT24yRHvUZay": {},
			"KT1LZFMGrdnPjRLsCZ1aEDUAF5myA5Eo4rQe": {},
			"KT1PDAELuX7CypUHinUgFgGFskKs7ytwh5Vw": {},
			"KT19xDbLsvQKnp9xqfDNPWJbKJJmV93dHDUa": {},
			"KT1Cz7TyVFvHxXpxLS57RFePrhTGisUpPhvD": {},
			"KT1LQ99RfGcmFe98PiBcGXuyjBkWzAcoXXhW": {},
			"KT1Gow8VzXZx3Akn5kvjACqnjnyYBxQpzSKr": {},
			"KT1DnfT4hfikoMY3uiPE9mQV4y3Xweramb2k": {},
			"KT1FuFDZGdw86p6krdBUKoZfEMkcUmezqX5o": {},
			"KT1SLWhfqPtQq7f4zLomh8BNgDeprF9B6d2M": {},
			"KT1THsDNgHtN56ew9VVCAUWnqPC81pqAxCEp": {},
			"KT1CM1g1o9RKDdtDKgcBWE59X2KgTc2TcYtC": {},
			"KT1W148mcjmfvr9J2RvWcGHxsAFApq9mcfgT": {},
			"KT1HvwFnXteMbphi7mfPDhCWkZSDvXEz8iyv": {},
			"KT1RUT25eGgo9KKWXfLhj1xYjghAY1iZ2don": {},
			"KT1EWLAQGPMF2uhtVRPaCH2vtFVN36Njdr6z": {},
			"KT1WPEis2WhAc2FciM2tZVn8qe6pCBe9HkDp": {},
			"KT1Msatnmdy24sQt6knzpALs4tvHfSPPduA2": {},
			"KT1A56dh8ivKNvLiLVkjYPyudmnY2Ti5Sba3": {},
			"KT1KRyTaxCAM3YRquifEe29BDbUKNhJ6hdtx": {},
			"KT1FL3C6t9Lyfskyb6rQrCRQTnf7M9t587VM": {},
			"KT1Q1kfbvzteafLvnGz92DGvkdypXfTGfEA3": {},
			"KT1CjfCztmRpsyUee1nLa9Wcpfr7vgwqRZmk": {},
			"KT1MHDHRLugz3A4qP6KqZDpa7FFmZfcJauV4": {},
			"KT1BvVxWM6cjFuJNet4R9m64VDCN2iMvjuGE": {},
			"KT1PyX9b8WmShQjqNgDQsvxqj9UYdmHLr3xg": {},
			"KT1XTXBsEauzcv3uPvVXW92mVqrx99UGsb9T": {},
			"KT1Puc9St8wdNoGtLiD2WXaHbWU7styaxYhD": {},
			"KT19c8n5mWrqpxMcR3J687yssHxotj88nGhZ": {},
			"KT1DrJV8vhkdLEj76h1H9Q4irZDqAkMPo1Qf": {},
			"KT1D68BvUm9N1fcq6uaZnyZvmBkBvj9biyPu": {},
			"KT1CT7S2b9hXNRxRrEcany9sak1qe4aaFAZJ": {},
			"KT1FHqsvc7vRS3u54L66DdMX4gb6QKqxJ1JW": {},
			"KT1QwBaLj5TRaGU3qkU4ZKKQ5mvNvyyzGBFv": {},
			"KT1TxqZ8QtKvLu3V3JH7Gx58n7Co8pgtpQU5": {},
			"KT1VqarPDicMFn1ejmQqqshUkUXTCTXwmkCN": {},
			"KT1AafHA1C1vk959wvHWBispY9Y2f3fxBUUo": {},
		},
	}
}

// Parse -
func (p *Jakarta) Parse(ctx context.Context, script noderpc.Script, old *modelsContract.Contract, previous, next protocol.Protocol, timestamp time.Time, tx models.Transaction) error {
	codeBytes, err := json.Marshal(script.Code)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := json.Compact(&buf, codeBytes); err != nil {
		return err
	}

	newHash, err := contract.ComputeHash(buf.Bytes())
	if err != nil {
		return err
	}

	var s bcd.RawScript
	if err := json.Unmarshal(buf.Bytes(), &s); err != nil {
		return err
	}

	contractScript := modelsContract.Script{
		Hash:      newHash,
		Code:      s.Code,
		Storage:   s.Storage,
		Parameter: s.Parameter,
		Views:     s.Views,
	}

	if err := tx.Scripts(ctx, &contractScript); err != nil {
		return err
	}

	old.JakartaID = contractScript.ID

	m := &migration.Migration{
		ContractID:     old.ID,
		Level:          next.StartLevel,
		ProtocolID:     next.ID,
		PrevProtocolID: previous.ID,
		Timestamp:      timestamp,
		Kind:           types.MigrationKindUpdate,
	}

	return tx.Migrations(ctx, m)
}

// IsMigratable -
func (p *Jakarta) IsMigratable(address string) bool {
	_, ok := p.contracts[address]
	return ok
}
