package resp

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type BigNumber struct {
	Value    string
	Negative bool
}

func NewBigNumber() Value {
	return &BigNumber{}
}

func (v *BigNumber) Read(r io.Reader) error {
	var strVal strings.Builder
	var currByte [1]byte
	negative := false
	seenSign := false

	for {
		_, err := io.ReadFull(r, currByte[:])
		if err != nil {
			return err
		}

		currRune := rune(currByte[0])
		switch {
		case currRune == '+' && strVal.Len() == 0 && !seenSign:
			negative = false
			seenSign = true
		case currRune == '-' && strVal.Len() == 0 && !seenSign:
			negative = true
			seenSign = true
		case currRune == '0':
			strVal.WriteByte(currByte[0])
		case currRune == '1':
			strVal.WriteByte(currByte[0])
		case currRune == '2':
			strVal.WriteByte(currByte[0])
		case currRune == '3':
			strVal.WriteByte(currByte[0])
		case currRune == '4':
			strVal.WriteByte(currByte[0])
		case currRune == '5':
			strVal.WriteByte(currByte[0])
		case currRune == '6':
			strVal.WriteByte(currByte[0])
		case currRune == '7':
			strVal.WriteByte(currByte[0])
		case currRune == '8':
			strVal.WriteByte(currByte[0])
		case currRune == '9':
			strVal.WriteByte(currByte[0])
		case currRune == '\r':
			_, err := io.ReadFull(r, currByte[:])
			if err != nil {
				return err
			}

			if rune(currByte[0]) == '\n' {
				v.Negative = negative
				v.Value = strVal.String()

				return nil
			}

			return errors.New("a BigNumber contained \r without \n")
		default:
			return fmt.Errorf("unexpected character in big number: %s", string(currByte[0]))
		}
	}
}

func (v *BigNumber) WriteTo(w io.Writer) (n int64, err error) {
	var written int
	if v.Negative {
		written, err = w.Write([]byte(fmt.Sprintf("(-%s\r\n", v.Value)))
	} else {
		written, err = w.Write([]byte(fmt.Sprintf("(%s\r\n", v.Value)))
	}

	return int64(written), err
}
