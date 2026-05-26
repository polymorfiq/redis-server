package resp

import (
	"errors"
	"io"
)

type Value interface {
	Read(io.Reader) error
	io.WriterTo
}

var encodings = map[string]func() Value{
	"#": NewBoolean,
	"_": NewNull,
	"+": NewSimpleString,
	":": NewInteger,
	"-": NewError,
	"$": NewBulkString,
	"!": NewBulkError,
	"*": NewArray,
	"~": NewSet,
	">": NewPush,
	",": NewDouble,
	"(": NewBigNumber,
	"=": NewVerbatimString,
	"%": NewMap,
	"|": NewAttributes,
}

func ReadValue(r io.Reader) (Value, error) {
	var typeByte [1]byte
	bytesRead, err := r.Read(typeByte[:])
	if err != nil {
		return nil, err
	} else if bytesRead != 1 {
		return nil, errors.New("could not read command byte")
	}

	valIdentStr := string(typeByte[0])
	valFunc, known := encodings[valIdentStr]
	if !known {
		return nil, errors.New("Unknown identifier " + valIdentStr)
	}

	val := valFunc()
	err = val.Read(r)
	if err != nil {
		return nil, err
	}

	return val, nil
}
