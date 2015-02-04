package arena

type Arena struct {
	buf []byte

	offset int
}

func NewArena(size int) *Arena {
	a := new(Arena)

	a.buf = make([]byte, size, size)
	a.offset = 0

	return a
}

func (a *Arena) Make(size int) []byte {
	if len(a.buf) < size || len(a.buf)-a.offset < size {
		return make([]byte, size)
	}

	b := a.buf[a.offset : size+a.offset]
	a.offset += size
	return b
}

func (a *Arena) Reset() {
	a.offset = 0
}
