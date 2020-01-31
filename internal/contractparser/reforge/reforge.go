package reforge

// func String(input string) string {
// 	1. Micheline
// 	2. UTF-8
// 	3. return input
// }

// Micheline -
func Micheline(hex string) string {
	// data = "0006"
	// map[string]interface{}
	// json.Marshal()
	// res := `{ "int": "6" }`
	// r := gjson.Parse(res)
	// s := MichelineToMichelson(r, false)

	// if len(hex) < 2 {
	// 	panic(hex)
	// }

	// var code string
	// var offset int

	// fieldType := hex[offset : offset+2]
	// offset += 2

	// switch fieldType {
	// case "00":
	// 	return hex
	// default:
	// 	panic(hex)
	// }

	return ""
}

// func decodeInt(hex string, offset int, signed bool) (string int) {
// 	var buffer string
// 	var i int

// 	for (offset + i * 2) < len(hex) {
// 		start := offset + i * 2
// 		end := start + 2
// 		part := hex[start:end]
// 		buffer += part
// 		i += 1
// 	}
// }
