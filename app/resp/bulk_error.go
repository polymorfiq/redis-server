package resp

import (
	"fmt"
	"io"
)

type BulkError struct {
	Contents []byte
}

func NewBulkError() Value {
	return &BulkError{}
}

func (v *BulkError) Read(r io.Reader) error {
	strContents := NewBulkString().(*BulkString)
	err := strContents.Read(r)
	if err != nil {
		return err
	}

	v.Contents = strContents.Contents

	return nil
}

func (v *BulkError) WriteTo(w io.Writer) (n int64, err error) {
	if v.Contents == nil {
		written, err := io.WriteString(w, fmt.Sprintf("!-1\r\n"))
		return int64(written), err
	}

	strVal := string(v.Contents)
	written, err := io.WriteString(w, fmt.Sprintf("!%d\r\n%s\r\n", len(v.Contents), strVal))
	return int64(written), err
}

func (v *BulkError) Error() string {
	return string(v.Contents)
}

func BulkErrorFromString(content string) *BulkError {
	err := NewBulkError().(*BulkError)
	err.Contents = []byte(content)
	return err
}

func NullBulkString() *BulkString {
	bulkStr := NewBulkString().(*BulkString)
	bulkStr.Contents = nil

	return bulkStr
}
