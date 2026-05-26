package resp

import (
	"fmt"
	"io"
)

type BulkString struct {
	Contents []byte
}

func NewBulkString() Value {
	return &BulkString{}
}

func BulkStringFromString(s string) *BulkString {
	bulkStr := NewBulkString().(*BulkString)
	bulkStr.Contents = []byte(s)
	return bulkStr
}

func (v *BulkString) Read(r io.Reader) error {
	strLength := NewInteger().(*Integer)
	err := strLength.Read(r)
	if err != nil {
		return err
	}

	if strLength.Value == -1 {
		v.Contents = nil
		return nil
	}

	strBytes := make([]byte, strLength.Value)
	_, err = io.ReadFull(r, strBytes)
	if err != nil {
		return err
	}

	var endBytes [2]byte
	_, err = io.ReadFull(r, endBytes[:])
	if err != nil {
		return err
	}

	if rune(endBytes[0]) != '\r' || rune(endBytes[1]) != '\n' {
		return fmt.Errorf("bulk string did not end with CRLF. Ended with %s", string(endBytes[:]))
	}

	v.Contents = strBytes
	return nil
}

func (v *BulkString) WriteTo(w io.Writer) (n int64, err error) {
	if v.Contents == nil {
		written, err := io.WriteString(w, fmt.Sprintf("$-1\r\n"))
		return int64(written), err
	}

	strVal := string(v.Contents)
	written, err := io.WriteString(w, fmt.Sprintf("$%d\r\n%s\r\n", len(v.Contents), strVal))
	return int64(written), err
}

func (v *BulkString) String() string {
	return string(v.Contents)
}
