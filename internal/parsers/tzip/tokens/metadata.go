package tokens

import (
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/tidwall/gjson"
)

const (
	pathTokenID = "args.0.int"
	pathMap     = "args.1.#.args"

	keySymbol             = "symbol"
	keyName               = "name"
	keyDecimals           = "decimals"
	keyDescr              = "description"
	keyArtifactURI        = "artifactUri"
	keyDisplayURI         = "displayUri"
	keyThumbnailURI       = "thumbnailUri"
	keyExternalURI        = "externalUri"
	keyIsTransferable     = "isTransferable"
	keyIsBooleanAmount    = "isBooleanAmount"
	keyShouldPreferSymbol = "shouldPreferSymbol"
	keyCreators           = "creators"
	keyTags               = "tags"
	keyFormats            = "formats"
)

// Empty key name
const (
	EmptyStringKey = "@@empty"
)

// TokenMetadata -
type TokenMetadata struct {
	Level              int64                  `json:"-"`
	Timestamp          time.Time              `json:"-"`
	TokenID            uint64                 `json:"-"`
	Symbol             string                 `json:"symbol"`
	Name               string                 `json:"name"`
	Decimals           *int64                 `json:"decimals"`
	Description        string                 `json:"description"`
	ArtifactURI        string                 `json:"artifactUri"`
	DisplayURI         string                 `json:"displayUri"`
	ThumbnailURI       string                 `json:"thumbnailUri"`
	ExternalURI        string                 `json:"externalUri"`
	IsTransferable     bool                   `json:"isTransferable"`
	IsBooleanAmount    bool                   `json:"isBooleanAmount"`
	ShouldPreferSymbol bool                   `json:"shouldPreferSymbol"`
	Creators           []string               `json:"creators"`
	Tags               []string               `json:"tags"`
	Formats            json.RawMessage        `json:"formats"`
	Extras             map[string]interface{} `json:"-"`

	Link string `json:"-"`
}

// ToModel -
func (m *TokenMetadata) ToModel(address, network string) tokenmetadata.TokenMetadata {
	return tokenmetadata.TokenMetadata{
		Network:            network,
		Contract:           address,
		Level:              m.Level,
		Timestamp:          m.Timestamp,
		TokenID:            m.TokenID,
		Symbol:             m.Symbol,
		Decimals:           m.Decimals,
		Name:               m.Name,
		Description:        m.Description,
		ArtifactURI:        m.ArtifactURI,
		DisplayURI:         m.DisplayURI,
		ThumbnailURI:       m.ThumbnailURI,
		ExternalURI:        m.ExternalURI,
		IsTransferable:     m.IsTransferable,
		IsBooleanAmount:    m.IsBooleanAmount,
		ShouldPreferSymbol: m.ShouldPreferSymbol,
		Creators:           m.Creators,
		Tags:               m.Tags,
		Formats:            types.Bytes(m.Formats),
		Extras:             m.Extras,
	}
}

func parseString(hexValue string) (string, error) {
	decoded, err := hex.DecodeString(hexValue)
	if err != nil {
		return "", err
	}

	if utf8.Valid(decoded) {
		return string(decoded), nil
	}
	return "", nil
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

	m.TokenID = tokenID.Uint()

	var err error
	m.Extras = make(map[string]interface{})
	for _, item := range arr.Array() {
		key := item.Get("0.string").String()
		value := item.Get("1.bytes").String()

		switch key {
		case "":
			m.Link = forge.DecodeString(value)
			m.Extras[EmptyStringKey] = m.Link
		case keySymbol:
			m.Symbol, err = parseString(value)
			if err != nil {
				return err
			}
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
			m.Name, err = parseString(value)
			if err != nil {
				return err
			}
		case keyArtifactURI:
			m.ArtifactURI, err = parseString(value)
			if err != nil {
				return err
			}
		case keyDescr:
			m.Description, err = parseString(value)
			if err != nil {
				return err
			}
		case keyDisplayURI:
			m.DisplayURI, err = parseString(value)
			if err != nil {
				return err
			}
		case keyThumbnailURI:
			m.ThumbnailURI, err = parseString(value)
			if err != nil {
				return err
			}
		case keyExternalURI:
			m.ExternalURI, err = parseString(value)
			if err != nil {
				return err
			}
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
	if second.Description != "" {
		m.Description = second.Description
	}
	if second.ArtifactURI != "" {
		m.ArtifactURI = second.ArtifactURI
	}
	if second.ExternalURI != "" {
		m.ExternalURI = second.ExternalURI
	}
	if second.DisplayURI != "" {
		m.DisplayURI = second.DisplayURI
	}
	if second.ThumbnailURI != "" {
		m.ThumbnailURI = second.ThumbnailURI
	}
	if second.IsBooleanAmount != m.IsBooleanAmount {
		m.IsBooleanAmount = second.IsBooleanAmount
	}
	if second.IsTransferable != m.IsTransferable {
		m.IsTransferable = second.IsTransferable
	}
	if second.ShouldPreferSymbol != m.ShouldPreferSymbol {
		m.ShouldPreferSymbol = second.ShouldPreferSymbol
	}
	if second.Creators != nil {
		m.Creators = second.Creators
	}
	if second.Tags != nil {
		m.Tags = second.Tags
	}
	if second.Formats != nil {
		m.Formats = second.Formats
	}

	for k, v := range second.Extras {
		m.Extras[k] = v
	}
}

func getStringArrayKey(data map[string]interface{}, keyName string) []string {
	if val, ok := data[keyName]; ok {
		delete(data, keyName)
		if s, ok := val.([]string); ok {
			return s
		}
	}
	return nil
}

func getStringKey(data map[string]interface{}, keyName string) string {
	if val, ok := data[keyName]; ok {
		delete(data, keyName)
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

func getBoolKey(data map[string]interface{}, keyName string) bool {
	if val, ok := data[keyName]; ok {
		delete(data, keyName)
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func getBytesKey(data map[string]interface{}, keyName string) json.RawMessage {
	if val, ok := data[keyName]; ok {
		delete(data, keyName)
		if b, ok := val.(json.RawMessage); ok {
			return b
		}
	}
	return nil
}

// UnmarshalJSON -
func (m *TokenMetadata) UnmarshalJSON(data []byte) error {
	res := make(map[string]interface{})
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}
	m.Name = getStringKey(res, keyName)
	m.Symbol = getStringKey(res, keySymbol)

	m.Description = getStringKey(res, keyDescr)
	m.ArtifactURI = getStringKey(res, keyArtifactURI)
	m.DisplayURI = getStringKey(res, keyDisplayURI)
	m.ThumbnailURI = getStringKey(res, keyThumbnailURI)
	m.ExternalURI = getStringKey(res, keyExternalURI)

	m.IsBooleanAmount = getBoolKey(res, keyIsBooleanAmount)
	m.IsTransferable = getBoolKey(res, keyIsTransferable)
	m.ShouldPreferSymbol = getBoolKey(res, keyShouldPreferSymbol)

	m.Creators = getStringArrayKey(res, keyCreators)
	m.Tags = getStringArrayKey(res, keyTags)

	m.Formats = getBytesKey(res, keyFormats)

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
