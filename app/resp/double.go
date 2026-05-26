package resp

import (
	"errors"
	"fmt"
	"io"
	"math"
)

type Double struct {
	IsSigned   bool
	Integral   int64
	Fractional uint64
	Exponent   int64
	Infinite   bool
	NaN        bool
}

func NewDouble() Value {
	return &Double{IsSigned: true}
}

type DoubleReadState int

const (
	ReadIntegralSign DoubleReadState = iota
	ReadIntegral
	ReadFractional
	ReadExponentSign
	ReadExponent
	ReadEndBytes
)

func (v *Double) Read(r io.Reader) error {

	var integralSign int64 = 1
	var integral int64
	var fractional int64
	var exponentSign int64 = 1
	var exponent int64

	var readState DoubleReadState
	if v.IsSigned {
		readState = ReadIntegralSign
	} else {
		readState = ReadIntegral
	}

	var currByte [1]byte
	parsingDouble := true
	for parsingDouble {
		_, err := io.ReadFull(r, currByte[:])
		if err != nil {
			return err
		}

		currRune := rune(currByte[0])
		if readState == ReadIntegralSign && currRune != '+' && currRune != '-' {
			readState = ReadIntegral
		} else if readState == ReadExponentSign && currRune != '+' && currRune != '-' {
			readState = ReadExponent
		}

		var val *int64
		switch readState {
		case ReadIntegralSign:
			val = &integralSign
		case ReadIntegral:
			val = &integral
		case ReadFractional:
			val = &fractional
		case ReadExponentSign:
			val = &exponentSign
		case ReadExponent:
			val = &exponent
		}

		switch {
		case readState == ReadIntegral && *val == 0 && currRune == 'i':
			var nfBytes [2]byte
			_, err := io.ReadFull(r, nfBytes[:])
			if err != nil {
				return err
			}

			if rune(nfBytes[0]) == 'n' && rune(nfBytes[1]) == 'f' {
				v.Integral = integralSign
				v.Infinite = true
				return nil
			} else {
				return fmt.Errorf("Unexpected 'i%s' in double", string(nfBytes[:]))
			}

		case readState == ReadIntegral && *val == 0 && currRune == 'n':
			var anBytes [2]byte
			_, err := io.ReadFull(r, anBytes[:])
			if err != nil {
				return err
			}

			if rune(anBytes[0]) == 'a' && rune(anBytes[1]) == 'n' {
				v.NaN = true
				return nil
			} else {
				return fmt.Errorf("Unexpected 'n%s' in double", string(anBytes[:]))
			}

		case currRune == '+':
			if readState == ReadIntegralSign {
				*val = 1
				readState = ReadIntegral
			} else if readState == ReadExponentSign {
				*val = 1
				readState = ReadExponent
			} else {
				return fmt.Errorf("Invalid + position (%s)", doubleReadStateStr(readState))
			}
		case currRune == '-':
			if readState == ReadIntegralSign {
				*val = -1
				readState = ReadIntegral
			} else if readState == ReadExponentSign {
				*val = -1
				readState = ReadExponent
			} else {
				return fmt.Errorf("Invalid - position (%s)", doubleReadStateStr(readState))
			}
		case currRune == '.' && readState == ReadIntegral:
			readState = ReadFractional
		case currRune == 'e' && readState == ReadIntegral:
			readState = ReadExponentSign
		case currRune == 'e' && readState == ReadFractional:
			readState = ReadExponentSign
		case currRune == '0':
			*val = *val * 10
		case currRune == '1':
			*val = (*val * 10) + 1
		case currRune == '2':
			*val = (*val * 10) + 2
		case currRune == '3':
			*val = (*val * 10) + 3
		case currRune == '4':
			*val = (*val * 10) + 4
		case currRune == '5':
			*val = (*val * 10) + 5
		case currRune == '6':
			*val = (*val * 10) + 6
		case currRune == '7':
			*val = (*val * 10) + 7
		case currRune == '8':
			*val = (*val * 10) + 8
		case currRune == '9':
			*val = (*val * 10) + 9
		case currRune == '\r':
			_, err := io.ReadFull(r, currByte[:])
			if err != nil {
				return err
			}

			if rune(currByte[0]) == '\n' {
				parsingDouble = false
			} else {
				return errors.New("Number did not end in \r\n")
			}

		default:
			return fmt.Errorf("Unexpected character in double (%s): %v", doubleReadStateStr(readState), rune(currByte[0]))
		}
	}

	var doubleVal float64

	v.Infinite = false
	v.NaN = false
	v.Integral = integral * integralSign
	v.Fractional = uint64(fractional)
	v.Exponent = exponent * exponentSign

	doubleVal = float64(integral * integralSign)
	fDigits := float64(len(fmt.Sprint(fractional)))
	doubleVal = doubleVal + (float64(fractional) / math.Pow(10, fDigits))
	doubleVal = doubleVal * math.Pow(10, float64(exponent*exponentSign))
	return nil
}

func (v *Double) WriteTo(w io.Writer) (n int64, err error) {
	doubleVal := float64(v.Integral)
	fDigits := float64(len(fmt.Sprint(v.Fractional)))
	doubleVal = doubleVal + (float64(v.Fractional) / math.Pow(10, fDigits))
	doubleVal = doubleVal * math.Pow(10, float64(v.Exponent))

	written, err := w.Write([]byte(fmt.Sprintf(",%e\r\n", doubleVal)))
	return int64(written), err
}

func doubleReadStateStr(state DoubleReadState) string {
	switch state {
	case ReadIntegralSign:
		return "ReadIntegralSign"
	case ReadIntegral:
		return "ReadIntegral"
	case ReadFractional:
		return "ReadFractional"
	case ReadExponentSign:
		return "ReadExponentSign"
	case ReadExponent:
		return "ReadExponent"
	case ReadEndBytes:
		return "ReadEndBytes"
	}

	return "Unknown State"
}
