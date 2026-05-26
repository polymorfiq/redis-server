package resp

import (
	"errors"
	"fmt"
	"io"
	"slices"
)

type VerbatimString struct {
	Encoding string
	Data     []byte
}

func NewVerbatimString() Value {
	return &VerbatimString{}
}

func (v *VerbatimString) Read(r io.Reader) error {
	dataLength := NewInteger().(*Integer)
	err := dataLength.Read(r)
	if err != nil {
		return err
	}

	var encodingBytes [3]byte
	_, err = io.ReadFull(r, encodingBytes[:])
	if err != nil {
		return err
	}

	var colonByte [1]byte
	_, err = io.ReadFull(r, colonByte[:])
	if err != nil {
		return err
	}

	if string(colonByte[:]) != ":" {
		return fmt.Errorf("expected colon but got %s", string(colonByte[:]))
	}

	dataBytes := make([]byte, dataLength.Value)
	_, err = io.ReadFull(r, dataBytes)
	if err != nil {
		return err
	}

	var endBytes [2]byte
	_, err = io.ReadFull(r, endBytes[:])
	if err != nil {
		return err
	}

	if rune(endBytes[0]) != '\r' || rune(endBytes[1]) != '\n' {
		return errors.New("Verbatim String did not end with CRLF")
	}

	v.Encoding = string(encodingBytes[:])
	v.Data = dataBytes

	return nil
}

func (v *VerbatimString) WriteTo(w io.Writer) (n int64, err error) {
	prefix := fmt.Sprintf("=%d\r\n%s:", len(v.Data), v.Encoding)
	data := slices.Concat([]byte(prefix), v.Data, []byte("\r\n"))

	written, err := w.Write(data)
	return int64(written), err
}
