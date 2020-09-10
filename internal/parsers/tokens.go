package parsers

import "github.com/baking-bad/bcdhub/internal/database"

// TokenKey -
type TokenKey struct {
	Address    string
	Network    string
	Entrypoint string
}

// TokenViews -
type TokenViews map[TokenKey]database.TokenViewImplementation

// NewTokenViews -
func NewTokenViews(db database.DB) (TokenViews, error) {
	tokens, err := db.GetTokens()
	if err != nil {
		return nil, err
	}

	views := make(TokenViews)
	for _, token := range tokens {
		if len(token.Metadata.Views) == 0 {
			continue
		}

		for _, view := range token.Metadata.Views {
			for _, implementation := range view.Implementations {
				for _, entrypoint := range implementation.MichelsonParameterView.Entrypoints {
					views[TokenKey{
						Address:    token.Contract,
						Network:    token.Network,
						Entrypoint: entrypoint,
					}] = implementation
				}
			}
		}
	}

	return views, nil
}

// Get -
func (views TokenViews) Get(address, network, entrypoint string) (database.TokenViewImplementation, bool) {
	view, ok := views[TokenKey{
		Address:    address,
		Network:    network,
		Entrypoint: entrypoint,
	}]
	return view, ok
}
