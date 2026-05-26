package resp

import (
	"io"
)

type Null struct{}

func NewNull() Value {
	return &Null{}
}

func (v *Null) Read(_ io.Reader) error {
	return nil
}

func (v *Null) WriteTo(w io.Writer) (n int64, err error) {
	written, err := w.Write([]byte("_\r\n"))
	return int64(written), err
}
