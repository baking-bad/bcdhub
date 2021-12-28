package main

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

var tokenContracts = []string{
	"KT1RJ6PbjHpwc3M5rw5s2Nbmefwbuwbdxton",
	"KT1K9gCRgaLRFKTErYt1wVxA3Frb9FjasjTV",
	"KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
	"KT1LN4LPSqTMS7Sd2CJw4bbDGRkMv2t68Fy9",
	"KT19at7rQUvyjxnZ2fBv7D9zc8rkyG7gAoU8",
	"KT1VYsVfmobT7rsMVivvZ4J8i3bPiqz12NaH",
}

func testAccounts(ctx *config.Context) {
	for _, address := range tokenContracts {
		balances, err := ctx.TokenBalances.GetHolders(types.Mainnet, address, 0)
		if err != nil {
			logger.Err(err)
			return
		}

		for i := range balances {
			path := fmt.Sprintf("account/mainnet/%s", balances[i].Account.Address)
			if err := request(path); err != nil {
				logger.Err(err)
			}
			if err := request(fmt.Sprintf("%s/token_balances", path)); err != nil {
				logger.Err(err)
			}
		}
	}
}
