package resp

import (
	"errors"
	"fmt"
	"io"
)

type Boolean struct {
	Value bool
}

func NewBoolean() Value {
	return &Boolean{}
}

func (v *Boolean) Read(req io.Reader) error {
	var booleanVal bool
	var boolByte [1]byte
	_, err := io.ReadFull(req, boolByte[:])
	if err != nil {
		return err
	}

	switch rune(boolByte[0]) {
	case 't':
		booleanVal = true
	case 'f':
		booleanVal = false
	default:
		return fmt.Errorf("unknown boolean value (%s)", string(boolByte[0]))
	}

	var endBytes [2]byte
	_, err = io.ReadFull(req, endBytes[:])
	if err != nil {
		return err
	}

	if rune(endBytes[0]) != '\r' || rune(endBytes[1]) != '\n' {
		return errors.New("boolean did not end with CRLF")
	}

	if rune(endBytes[0]) != '\r' || rune(endBytes[1]) != '\n' {
		return errors.New("boolean did not end with CRLF")
	}

	v.Value = booleanVal
	return nil
}

func (v *Boolean) WriteTo(w io.Writer) (n int64, err error) {
	var written int
	if v.Value {
		written, err = w.Write([]byte("#t\r\n"))
	} else {
		written, err = w.Write([]byte("#f\r\n"))
	}

	return int64(written), err
}
