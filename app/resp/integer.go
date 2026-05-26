package resp

import (
	"errors"
	"fmt"
	"io"
)

type Integer struct {
	IsSigned bool
	Value    int64
}

func NewInteger() Value {
	return &Integer{IsSigned: true}
}

func (v *Integer) Read(r io.Reader) error {
	var val int64
	var currByte [1]byte
	var bytesRead int
	for {
		_, err := io.ReadFull(r, currByte[:])
		if err != nil {
			return err
		}

		currRune := rune(currByte[0])

		switch {
		case v.IsSigned && currRune == '+':
			v.IsSigned = true
			if bytesRead > 0 {
				return errors.New("number contained '+' outside of prefix")
			}
		case v.IsSigned && currRune == '-':
			if bytesRead == 0 {
				v.IsSigned = true
				val = -val
			} else {
				return errors.New("number contained '+' outside of prefix")
			}
		case currRune == '0':
			val = val * 10
		case currRune == '1':
			val = (val * 10) + 1
		case currRune == '2':
			val = (val * 10) + 2
		case currRune == '3':
			val = (val * 10) + 3
		case currRune == '4':
			val = (val * 10) + 4
		case currRune == '5':
			val = (val * 10) + 5
		case currRune == '6':
			val = (val * 10) + 6
		case currRune == '7':
			val = (val * 10) + 7
		case currRune == '8':
			val = (val * 10) + 8
		case currRune == '9':
			val = (val * 10) + 9
		case currRune == '\r':
			_, err := io.ReadFull(r, currByte[:])
			if err != nil {
				return err
			}

			if rune(currByte[0]) == '\n' {
				v.Value = val
				return nil
			}

			return errors.New("number did not end in \r\n")

		default:
			return fmt.Errorf("unexpected character in number: '%s'", string(rune(currByte[0])))
		}

		bytesRead++
	}
}

func (v *Integer) WriteTo(w io.Writer) (n int64, err error) {
	written, err := io.WriteString(w, fmt.Sprintf(":%d\r\n", v.Value))
	return int64(written), err
}

func IntegerFromInt(n int64) *Integer {
	integer := NewInteger().(*Integer)
	integer.IsSigned = n < 0
	integer.Value = n

	return integer
}
