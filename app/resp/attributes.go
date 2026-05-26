package resp

import (
	"fmt"
	"io"
)

type Attributes struct {
	Values map[interface{}]interface{}
}

func NewAttributes() Value {
	return &Attributes{}
}

func (v *Attributes) Read(r io.Reader) error {
	mapContents := NewMap().(*Map)
	err := mapContents.Read(r)
	if err != nil {
		return err
	}

	v.Values = mapContents.Values

	return nil
}

func (v *Attributes) WriteTo(w io.Writer) (n int64, err error) {
	for key, val := range v.Values {
		_, keyIsValue := key.(Value)
		if !keyIsValue {
			return 0, fmt.Errorf("key is not a valid RESP Value (%T): %v", key, key)
		}
		_, valIsValue := val.(Value)
		if !valIsValue {
			return 0, fmt.Errorf("value is not a valid RESP Value (%T): %v", key, key)
		}
	}

	written, err := w.Write([]byte(fmt.Sprintf("|%d\r\n", len(v.Values))))
	n = int64(written)
	if err != nil {
		return n, err
	}

	for key, val := range v.Values {
		key := key.(Value)
		val := val.(Value)

		keyBytes, err := key.WriteTo(w)
		n += keyBytes
		if err != nil {
			return n, err
		}

		valBytes, err := val.WriteTo(w)
		n += valBytes
		if err != nil {
			return n, err
		}
	}

	return n, nil
}
