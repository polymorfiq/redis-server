package resp

import (
	"fmt"
	"io"
	"strings"
)

type Array struct {
	Values []Value
	IsNull bool
}

func NewArray() Value {
	return &Array{}
}

func (v *Array) Read(r io.Reader) error {
	arrayLength := NewInteger().(*Integer)
	err := arrayLength.Read(r)
	if err != nil {
		return err
	}

	if arrayLength.Value == -1 {
		v.Values = nil
		return nil
	}

	arrayVal := make([]Value, arrayLength.Value)
	for idx := range arrayLength.Value {
		val, err := ReadValue(r)
		if err != nil {
			return fmt.Errorf("Error reading array element %d: %v", idx, err)
		}

		arrayVal[idx] = val
	}

	v.Values = arrayVal
	return nil
}

func (v *Array) WriteTo(w io.Writer) (n int64, err error) {
	if v.IsNull {
		written, err := io.WriteString(w, fmt.Sprintf("*-1\r\n"))
		return int64(written), err
	}

	written, err := io.WriteString(w, fmt.Sprintf("*%d\r\n", len(v.Values)))
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

func ArrayFromValues(values []Value) *Array {
	array := NewArray().(*Array)
	array.Values = values
	return array
}

func MaybeSimpleString(s string) Value {
	if strings.Contains(s, "\r") || strings.Contains(s, "\n") {
		return BulkStringFromString(s)
	}

	return SimpleStringFromString(s)
}

func NullArray() Value {
	null := NewArray().(*Array)
	null.IsNull = true
	return null
}

func ArrayOfStrings(strs []string) *Array {
	array := NewArray().(*Array)
	for _, str := range strs {
		array.Values = append(array.Values, MaybeSimpleString(str))
	}

	return array
}

func ArrayOfSimpleStrings(strs []string) *Array {
	array := NewArray().(*Array)
	for _, str := range strs {
		array.Values = append(array.Values, SimpleStringFromString(str))
	}

	return array
}
