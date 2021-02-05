package pack

import (
	"fmt"
	"testing"

	"github.com/tidwall/gjson"
)

func TestMicheline(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "bytes",
			input: `{
				"bytes": "000018896fcfc6690baefa9aedc6d759f9bf05727e8c"
			}`,
			expected: "050a00000016000018896fcfc6690baefa9aedc6d759f9bf05727e8c",
		},
		{
			name: "string",
			input: `{
				"string": "Game one!"
			}`,
			expected: "05010000000947616d65206f6e6521",
		},
		{
			name: "int",
			input: `{
				"int": "505506"
			}`,
			expected: "0500a2da3d",
		},
		{
			name: "string 2",
			input: `{
				"string": "sMeta"
			}`,
			expected: "050100000005734d657461",
		},
		{
			name: "bls12_381_g1",
			input: `{
				"bytes": "063bd6e11e2fcaac1dd8cf68c6b1925a73c3c583e298ed37c41c3715115cf96358a42dbe85a0228cbfd8a6c8a8c54cd015b5ae2860d1cc47f84698d951f14d9448d03f04df2ca0ffe609a2067d6f1a892163a5e05e541279134cae52b1f23c6b"
			}`,
			expected: "050a00000060063bd6e11e2fcaac1dd8cf68c6b1925a73c3c583e298ed37c41c3715115cf96358a42dbe85a0228cbfd8a6c8a8c54cd015b5ae2860d1cc47f84698d951f14d9448d03f04df2ca0ffe609a2067d6f1a892163a5e05e541279134cae52b1f23c6b",
		},
		{
			name: "bls12_381_g1",
			input: `{
				"bytes": "11f5b5db1da7f1f26217edcce2219d016003af6e5b4d1ca3ad0ff477e354717e658bf16beddc4f4fb76ce39d3327811e0601709dc7ed98c70463cfa1ba33f99851b52b51d1a042d7425bec6277287441c399973632445ce61e7fdd63a70f0f60"
			}`,
			expected: "050a0000006011f5b5db1da7f1f26217edcce2219d016003af6e5b4d1ca3ad0ff477e354717e658bf16beddc4f4fb76ce39d3327811e0601709dc7ed98c70463cfa1ba33f99851b52b51d1a042d7425bec6277287441c399973632445ce61e7fdd63a70f0f60",
		},
		{
			name: "bls12_381_g2",
			input: `{
				"bytes": "10c6d5cdca84fc3c7f33061add256f48e0ab03a697832b338901898b650419eb6f334b28153fb73ad2ecd1cd2ac67053161e9f46cfbdaf7b1132a4654a55162850249650f9b873ac3113fa8c02ef1cd1df481480a4457f351d28f4da89d19fa405c3d77f686dc9a24d2681c9184bf2b091f62e6b24df651a3da8bd7067e14e7908fb02f8955b84af5081614cb5bc49b416d9edf914fc608c441b3f2eb8b6043736ddb9d4e4d62334a23b5625c14ef3e1a7e99258386310221b22d83a5eac035c"
			}`,
			expected: "050a000000c010c6d5cdca84fc3c7f33061add256f48e0ab03a697832b338901898b650419eb6f334b28153fb73ad2ecd1cd2ac67053161e9f46cfbdaf7b1132a4654a55162850249650f9b873ac3113fa8c02ef1cd1df481480a4457f351d28f4da89d19fa405c3d77f686dc9a24d2681c9184bf2b091f62e6b24df651a3da8bd7067e14e7908fb02f8955b84af5081614cb5bc49b416d9edf914fc608c441b3f2eb8b6043736ddb9d4e4d62334a23b5625c14ef3e1a7e99258386310221b22d83a5eac035c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := gjson.Parse(tt.input)
			result, err := Micheline(parsed)
			if err != nil {
				t.Errorf("error in Micheline, error: %v", err)
			}
			if fmt.Sprintf("%x", result) != tt.expected {
				t.Errorf("error in Micheline, got: %x, expected: %v", result, tt.expected)
			}
		})
	}
}
