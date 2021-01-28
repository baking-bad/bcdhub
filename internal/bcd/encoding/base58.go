package encoding

import (
	"encoding/hex"
	"errors"

	"github.com/ebellocchia/go-base58"
)

var base58Enc = base58.New(base58.AlphabetBitcoin)

type base58Encoding struct {
	EncodedPrefix []byte
	EncodedLength int
	DecodedPrefix []byte
	DecodedLength int
	DataType      string
}

var base58Encodings = []base58Encoding{
	{[]byte("B"), 51, []byte{1, 52}, 32, "block hash"},
	{[]byte("o"), 51, []byte{5, 116}, 32, "operation hash"},
	{[]byte("Lo"), 52, []byte{133, 233}, 32, "operation list hash"},
	{[]byte("LLo"), 53, []byte{29, 159, 109}, 32, "operation list list hash"},
	{[]byte("P"), 51, []byte{2, 170}, 32, "protocol hash"},
	{[]byte("Co"), 52, []byte{79, 199}, 32, "context hash"},

	{[]byte("tz1"), 36, []byte{6, 161, 159}, 20, "ed25519 public key hash"},
	{[]byte("tz2"), 36, []byte{6, 161, 161}, 20, "secp256k1 public key hash"},
	{[]byte("tz3"), 36, []byte{6, 161, 164}, 20, "p256 public key hash"},
	{[]byte("KT1"), 36, []byte{2, 90, 121}, 20, "Originated address"},

	{[]byte("expr"), 54, []byte{13, 44, 64, 27}, 32, "script expression"},
	{[]byte("edsk"), 54, []byte{13, 15, 58, 7}, 32, "ed25519 seed"},
	{[]byte("edpk"), 54, []byte{13, 15, 37, 217}, 32, "ed25519 public key"},
	{[]byte("spsk"), 54, []byte{17, 162, 224, 201}, 32, "secp256k1 secret key"},
	{[]byte("p2sk"), 54, []byte{16, 81, 238, 189}, 32, "p256 secret key"},

	{[]byte("sppk"), 55, []byte{3, 254, 226, 86}, 33, "secp256k1 public key"},
	{[]byte("p2pk"), 55, []byte{3, 178, 139, 127}, 33, "p256 public key"},
	{[]byte("SSp"), 53, []byte{38, 248, 136}, 33, "secp256k1 scalar"},
	{[]byte("GSp"), 53, []byte{5, 92, 0}, 33, "secp256k1 element"},

	{[]byte("edsk"), 98, []byte{43, 246, 78, 7}, 64, "ed25519 secret key"},
	{[]byte("edsig"), 99, []byte{9, 245, 205, 134, 18}, 64, "ed25519 signature"},
	{[]byte("spsig"), 99, []byte{13, 115, 101, 19, 63}, 64, "secp256k1 signature"},
	{[]byte("p2sig"), 98, []byte{54, 240, 44, 52}, 64, "p256 signature"},
	{[]byte("sig"), 96, []byte{4, 130, 43}, 64, "generic signature"},

	{[]byte("Net"), 15, []byte{87, 82, 0}, 4, "chain id"},

	{[]byte("id"), 30, []byte{153, 103}, 16, "cryptobox public key hash"},

	{[]byte("edesk"), 88, []byte{7, 90, 60, 179, 41}, 56, "ed25519 encrypted seed"},
	{[]byte("spesk"), 88, []byte{9, 237, 241, 174, 150}, 56, "secp256k1 encrypted secret key"},
	{[]byte("p2esk"), 88, []byte{9, 48, 57, 115, 171}, 56, "p256_encrypted_secret_key"},
}

func getBase58EncodingForDecode(data []byte) (base58Encoding, error) {
	for _, e := range base58Encodings {
		if len(data) != e.DecodedLength+len(e.DecodedPrefix) {
			continue
		}
		found := true
		for i := range e.DecodedPrefix {
			if e.DecodedPrefix[i] != data[i] {
				found = false
				break
			}
		}
		if found {
			return e, nil
		}
	}
	return base58Encoding{}, errors.New("Unknown base58 encoding")
}

func getBase58EncodingForEncode(data, prefix []byte) (base58Encoding, error) {
	for _, e := range base58Encodings {
		if len(data) != e.DecodedLength {
			continue
		}
		found := true
		for i := range prefix {
			if e.DecodedPrefix[i] != prefix[i] {
				found = false
				break
			}
		}
		if found {
			return e, nil
		}
	}
	return base58Encoding{}, errors.New("Unknown base58 encoding")
}

// DecodeBase58 -
func DecodeBase58(data string) ([]byte, error) {
	decoded, err := base58Enc.CheckDecode(data)
	if err != nil {
		return nil, err
	}
	enc, err := getBase58EncodingForDecode(decoded)
	if err != nil {
		return nil, err
	}

	return decoded[len(enc.DecodedPrefix):], nil
}

// DecodeBase58ToString -
func DecodeBase58ToString(data string) (string, error) {
	decoded, err := base58Enc.CheckDecode(data)
	if err != nil {
		return "", err
	}
	enc, err := getBase58EncodingForDecode(decoded)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(decoded[len(enc.DecodedPrefix):]), nil
}

// DecodeBase58String -
func DecodeBase58String(data string) (string, error) {
	b, err := DecodeBase58(data)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// EncodeBase58 -
func EncodeBase58(data, prefix []byte) (string, error) {
	enc, err := getBase58EncodingForEncode(data, prefix)
	if err != nil {
		return "", err
	}
	b := append(enc.DecodedPrefix, data...)

	return base58Enc.CheckEncode(b), nil
}

// EncodeBase58String -
func EncodeBase58String(data string, prefix []byte) (string, error) {
	b, err := hex.DecodeString(data)
	if err != nil {
		return "", err
	}
	return EncodeBase58(b, prefix)
}
