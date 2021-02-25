package tokens

import (
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/tidwall/gjson"
)

const (
	pathTokenID = "args.0.int"
	pathMap     = "args.1.#.args"

	keySymbol   = "symbol"
	keyName     = "name"
	keyDecimals = "decimals"
)

// Empty key name
const (
	EmptyStringKey = "@@empty"
)

// TokenMetadata -
type TokenMetadata struct {
	Level     int64
	Timestamp time.Time
	TokenID   int64
	Symbol    string
	Name      string
	Decimals  *int64
	Extras    map[string]interface{}

	Link string
}

// ToModel -
func (m *TokenMetadata) ToModel(address, network string) tokenmetadata.TokenMetadata {
	return tokenmetadata.TokenMetadata{
		Network:   network,
		Contract:  address,
		Level:     m.Level,
		Timestamp: m.Timestamp,
		TokenID:   m.TokenID,
		Symbol:    m.Symbol,
		Decimals:  m.Decimals,
		Name:      m.Name,
		Extras:    m.Extras,
	}
}

// Parse -
func (m *TokenMetadata) Parse(value gjson.Result, address string, ptr int64) error {
	if value.Get("prim").String() != consts.Pair {
		return ErrInvalidStorageStructure
	}
	arr := value.Get(pathMap)
	if !arr.IsArray() {
		return ErrInvalidStorageStructure
	}
	tokenID := value.Get(pathTokenID)
	if !tokenID.Exists() {
		return ErrInvalidStorageStructure
	}

	m.TokenID = tokenID.Int()

	m.Extras = make(map[string]interface{})
	for _, item := range arr.Array() {
		key := item.Get("0.string").String()
		value := item.Get("1.bytes").String()

		switch key {
		case "":
			m.Link = forge.DecodeString(value)
			m.Extras[EmptyStringKey] = m.Link
		case keySymbol:
			decoded, err := hex.DecodeString(value)
			if err != nil {
				return err
			}
			m.Symbol = string(decoded)
		case keyDecimals:
			b, err := hex.DecodeString(value)
			if err != nil {
				return err
			}
			decoded, err := strconv.ParseInt(string(b), 10, 64)
			if err != nil {
				return err
			}
			m.Decimals = &decoded
		case keyName:
			decoded, err := hex.DecodeString(value)
			if err != nil {
				return err
			}
			m.Name = string(decoded)
		default:
			m.Extras[key] = forge.DecodeString(value)
		}
	}
	return nil
}

// Merge -
func (m *TokenMetadata) Merge(second *TokenMetadata) {
	if second.Decimals != nil {
		m.Decimals = second.Decimals
	}
	if second.Symbol != "" {
		m.Symbol = second.Symbol
	}
	if second.Name != "" {
		m.Name = second.Name
	}
	for k, v := range second.Extras {
		m.Extras[k] = v
	}
}

// UnmarshalJSON -
func (m *TokenMetadata) UnmarshalJSON(data []byte) error {
	res := make(map[string]interface{})
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}
	if val, ok := res[keyName]; ok {
		if name, ok := val.(string); ok {
			m.Name = name
		}
		delete(res, keyName)
	}
	if val, ok := res[keySymbol]; ok {
		if symbol, ok := val.(string); ok {
			m.Symbol = symbol
		}
		delete(res, keySymbol)
	}
	if val, ok := res[keyDecimals]; ok {
		switch decimals := val.(type) {
		case float64:
			int64Val := int64(decimals)
			m.Decimals = &int64Val
		case int64:
			m.Decimals = &decimals
		case string:
			int64Val, err := strconv.ParseInt(decimals, 10, 64)
			if err != nil {
				logger.Errorf("TokenMetadata decimal Unmarshal error with string. Got %##v %T", res[keyDecimals], val)
			} else {
				m.Decimals = &int64Val
			}
		default:
			logger.Errorf("TokenMetadata decimal Unmarshal error. Wanted float64, int64 or (>_<) string, got %##v %T", res[keyDecimals], val)
		}
		delete(res, keyDecimals)
	}

	m.Extras = make(map[string]interface{})
	for key, value := range res {
		m.Extras[key] = value
	}
	return nil
}
