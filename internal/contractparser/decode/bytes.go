package decode

func decodeBytes(h string) (string, int) {
	return h[8:], len(h[8:])*2 + 8
}
