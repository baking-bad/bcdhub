package contractparser

func filterAnnotations(annots []string) []string {
	var ret []string

	for _, a := range annots {
		if len(a) < 2 {
			continue
		}

		if !isValidPrefix(a[0]) {
			continue
		}

		if !isChar(a[1]) {
			continue
		}

		ret = append(ret, a)
	}

	return ret
}

func isValidPrefix(c byte) bool {
	return c == '%' || c == '@' || c == ':'
}

func isChar(c byte) bool {
	return (c >= 65 && c <= 90) || (c >= 97 && c <= 122)
}
