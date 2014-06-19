package log

import (
	"io"
)

type Handler interface {
	Write(p []byte) (n int, err error)
	Close() error
}

type StreamHandler struct {
	w io.Writer
}

func NewStreamHandler(w io.Writer) (*StreamHandler, error) {
	h := new(StreamHandler)

	h.w = w

	return h, nil
}

func (h *StreamHandler) Write(b []byte) (n int, err error) {
	return h.w.Write(b)
}

func (h *StreamHandler) Close() error {
	return nil
}

type NullHandler struct {
}

func NewNullHandler() (*NullHandler, error) {
	return new(NullHandler), nil
}

func (h *NullHandler) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (h *NullHandler) Close() {

}
