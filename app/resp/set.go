package resp

import (
	"fmt"
	"io"
)

type Set struct {
	Values []Value
}

func NewSet() Value {
	return &Set{}
}

func (v *Set) Read(r io.Reader) error {
	values := NewArray().(*Array)
	err := values.Read(r)
	if err != nil {
		return err
	}

	v.Values = values.Values
	return nil
}

func (v *Set) WriteTo(w io.Writer) (n int64, err error) {
	if v.Values == nil {
		written, err := io.WriteString(w, fmt.Sprintf("~-1\r\n"))
		return int64(written), err
	}

	written, err := io.WriteString(w, fmt.Sprintf("~%d\r\n", len(v.Values)))
	n = int64(written)
	for _, val := range v.Values {
		valBytes, err := val.WriteTo(w)
		n += valBytes
		if err != nil {
			return n, err
		}
	}

	return n, nil
}
