package contractparser

import "strconv"

func filterAnnotations(annots []string) []string {
	var ret []string

	for _, a := range annots {
		if a[0] == '%' && isDigit(a[1:]) {
			continue
		}

		if !hasChar(a) {
			continue
		}

		ret = append(ret, a)
	}

	return ret
}

func isDigit(input string) bool {
	_, err := strconv.ParseUint(input, 10, 32)
	return err == nil
}

func hasChar(input string) bool {
	for _, c := range input {
		// check if char between A..Z or a..z
		if (c >= 65 && c <= 90) || (c >= 97 && c <= 122) {
			return true
		}
	}
	return false
}
