package resp

import (
	"fmt"
	"io"
	"strings"
)

type Error struct {
	Contents string
}

func NewError() Value {
	return &Error{}
}

func (v *Error) Read(r io.Reader) error {
	contentsStr := NewSimpleString().(*SimpleString)
	err := contentsStr.Read(r)
	if err != nil {
		return err
	}

	v.Contents = contentsStr.Contents
	return nil
}

func (v *Error) WriteTo(w io.Writer) (n int64, err error) {
	written, err := io.WriteString(w, fmt.Sprintf("-%s\r\n", v.Contents))
	return int64(written), err
}

func (v *Error) Error() string {
	return v.Contents
}

func ErrorFromString(content string) *Error {
	content = strings.ReplaceAll(content, "\r", "")
	content = strings.ReplaceAll(content, "\n", "")

	err := NewError().(*Error)
	err.Contents = content
	return err
}
