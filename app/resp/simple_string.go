package resp

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type SimpleString struct {
	Contents string
}

func NewSimpleString() Value {
	return &SimpleString{}
}

func (v *SimpleString) Read(r io.Reader) error {
	var strVal strings.Builder
	var currByte [1]byte

	for {
		_, err := io.ReadFull(r, currByte[:])
		if err != nil {
			return err
		}

		curRune := rune(currByte[0])
		switch {
		case curRune == '\r':
			_, err := io.ReadFull(r, currByte[:])
			if err != nil {
				return err
			}

			if rune(currByte[0]) == '\n' {
				v.Contents = strVal.String()
				return nil
			} else {
				return errors.New("Simple String contained \r without \n")
			}
		case curRune == '\n':
			return errors.New("Simple String contained \n without \r")

		default:
			strVal.WriteByte(currByte[0])
		}
	}
}

func (v *SimpleString) WriteTo(w io.Writer) (n int64, err error) {
	written, err := io.WriteString(w, fmt.Sprintf("+%s\r\n", v.Contents))
	return int64(written), err
}

func (v *SimpleString) String() string {
	return v.Contents
}

func SimpleStringFromString(s string) *SimpleString {
	bulkStr := NewSimpleString().(*SimpleString)
	bulkStr.Contents = s
	return bulkStr
}
