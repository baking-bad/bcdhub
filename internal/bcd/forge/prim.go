package forge

import (
	"fmt"

	"github.com/pkg/errors"
)

type prim struct {
	Value string
}

func newPrim(p string) *prim {
	return &prim{
		Value: p,
	}
}

// Unforge -
func (p *prim) Unforge(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, errors.Wrap(ErrTooFewBytes, "prim.Unforge: 0")
	}
	key := int(data[0])
	if key > len(primKeywords) {
		return 0, errors.Wrap(ErrInvalidKeyword, fmt.Sprintf("prim.Unforge: %d", key))
	}
	p.Value = primKeywords[key]
	return 1, nil
}

// Forge -
func (p *prim) Forge() ([]byte, error) {
	for i := range primKeywords {
		if primKeywords[i] == p.Value {
			return []byte{byte(i)}, nil
		}
	}
	return nil, errors.Wrap(ErrInvalidKeyword, p.Value)
}
