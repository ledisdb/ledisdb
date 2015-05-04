package server

func lowerSlice(buf []byte) []byte {
	for i, r := range buf {
		if 'A' <= r && r <= 'Z' {
			r += 'a' - 'A'
		}

		buf[i] = r
	}
	return buf
}

func upperSlice(buf []byte) []byte {
	for i, r := range buf {
		if 'a' <= r && r <= 'z' {
			r -= 'a' - 'A'
		}

		buf[i] = r
	}
	return buf
}
