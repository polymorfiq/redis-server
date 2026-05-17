package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"strings"
)

var _ = net.Listen
var _ = os.Exit

func main() {
	l, err := net.Listen("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	fmt.Printf("New connection from %v\n", conn.RemoteAddr())

	for {
		val, err := readRespValue(conn)
		if err != nil {
			log.Println("Error reading RESP value: ", err.Error())
			break
		}

		err = handleReq(val, conn)
		if err != nil {
			log.Println("Error handling request: ", err.Error())
			break
		}

		//line, err := bufio.NewReader(conn).ReadString('\n')
		//if err != nil {
		//	fmt.Printf("Error reading from connection (%s): %s\n", conn.RemoteAddr(), err.Error())
		//}
		//
		//line = strings.TrimSuffix(line, "\n")
		//fmt.Printf("MSG (%s): %s\n", conn.RemoteAddr(), line)
		//
		//_, err = conn.Write([]byte("+PONG\r\n"))
		//if err != nil {
		//	fmt.Printf("Error writing to connection (%s): %s\n", conn.RemoteAddr(), err.Error())
		//}
	}
}

func readRespValue(req io.Reader) (interface{}, error) {
	var typeByte [1]byte
	bytesRead, err := req.Read(typeByte[:])
	if err != nil {
		return nil, err
	} else if bytesRead != 1 {
		return nil, errors.New("Could not read command byte")
	}

	cmdRune := rune(typeByte[0])
	switch cmdRune {
	case '#':
		return readBoolean(req)
	case '_':
		return readNull(req)
	case '+':
		return readSimpleString(req)
	case ':':
		return readInteger(req, true)
	case '-':
		errorStr, err := readSimpleString(req)
		if err != nil {
			return nil, err
		}

		return errors.New(errorStr), nil
	case '$':
		strPtr, err := readBulkString(req)
		if err != nil {
			return nil, err
		}

		if strPtr == nil {
			return nil, nil
		}

		return string(*strPtr), nil
	case '!':
		errorPtr, err := readBulkString(req)
		if err != nil {
			return nil, err
		}

		if errorPtr == nil {
			return nil, nil
		}

		return errors.New(string(*errorPtr)), nil
	case '*':
		return readArray(req)
	case '~':
		return readSet(req)
	case '>':
		return readPush(req)
	case ',':
		return readDouble(req, true)
	case '(':
		return readBigNumber(req)
	case '=':
		return readVerbatimString(req)
	case '%':
		return readMap(req)
	case '|':
		return readAttributes(req)
	default:
		return nil, fmt.Errorf("Unknown command (%s)", string(cmdRune))
	}
}

func writeSimpleString(req io.Writer, str string) error {
	_, err := io.WriteString(req, fmt.Sprintf("+%s\r\n", str))
	if err != nil {
		return err
	}

	return nil
}

func readSimpleString(req io.Reader) (string, error) {
	var strVal strings.Builder
	var currByte [1]byte

	for {
		_, err := io.ReadFull(req, currByte[:])
		if err != nil {
			return "", err
		}

		curRune := rune(currByte[0])
		switch {
		case curRune == '\r':
			_, err := io.ReadFull(req, currByte[:])
			if err != nil {
				return "", err
			}

			if rune(currByte[0]) == '\n' {
				return strVal.String(), nil
			} else {
				return "", errors.New("Simple String contained \r without \n")
			}
		case curRune == '\n':
			return "", errors.New("Simple String contained \n without \r")
		default:
			strVal.WriteByte(currByte[0])
		}
	}
}

func writeBulkString(req io.Writer, str *string) (err error) {
	if str == nil {
		_, err = io.WriteString(req, "$-1\r\n")
	} else {
		strLen := int64(len(*str))
		_, err = io.WriteString(req, fmt.Sprintf("$%d\r\n%s\r\n", strLen, *str))
	}

	return
}

func readBulkString(req io.Reader) (*[]byte, error) {
	strLength, err := readInteger(req, false)
	if err != nil {
		return nil, err
	}

	if strLength == -1 {
		return nil, nil
	}

	strBytes := make([]byte, strLength)
	_, err = io.ReadFull(req, strBytes)
	if err != nil {
		return nil, err
	}

	var endBytes [2]byte
	_, err = io.ReadFull(req, endBytes[:])
	if err != nil {
		return nil, err
	}

	if rune(endBytes[0]) != '\r' || rune(endBytes[1]) != '\n' {
		return nil, errors.New("Bulk String did not end with CRLF")
	}

	return &strBytes, nil
}

func readArray(req io.Reader) ([]interface{}, error) {
	arrayLength, err := readInteger(req, true)
	if err != nil {
		return nil, err
	}

	if arrayLength == -1 {
		return nil, nil
	}

	arrayVal := make([]interface{}, arrayLength)
	for idx := range arrayLength {
		val, err := readRespValue(req)
		if err != nil {
			return nil, fmt.Errorf("Error reading array element %d: %v", idx, err)
		}

		arrayVal[idx] = val
	}

	return arrayVal, nil
}

type Set struct {
	items []interface{}
}

func readSet(req io.Reader) (Set, error) {
	data, err := readArray(req)
	if err != nil {
		return Set{}, err
	}

	return Set{items: data}, nil
}

type Push struct {
	items []interface{}
}

func readPush(req io.Reader) (Push, error) {
	data, err := readArray(req)
	if err != nil {
		return Push{}, err
	}

	return Push{items: data}, nil
}

func readMap(req io.Reader) (map[interface{}]interface{}, error) {
	numEntries, err := readInteger(req, true)
	if err != nil {
		return nil, err
	}

	mapVal := make(map[interface{}]interface{}, numEntries)
	for idx := range numEntries {
		key, err := readRespValue(req)
		if err != nil {
			return nil, fmt.Errorf("Error reading map key %d: %v", idx, err)
		}

		val, err := readRespValue(req)
		if err != nil {
			return nil, fmt.Errorf("Error reading map val %d: %v", idx, err)
		}

		mapVal[key] = val
	}

	return mapVal, nil
}

type Attributes struct {
	attributes map[interface{}]interface{}
}

func readAttributes(req io.Reader) (Attributes, error) {
	attribs, err := readMap(req)
	if err != nil {
		return Attributes{}, err
	}

	return Attributes{attributes: attribs}, nil
}

func readBoolean(req io.Reader) (bool, error) {
	var booleanVal bool
	var boolByte [1]byte
	_, err := io.ReadFull(req, boolByte[:])
	if err != nil {
		return false, err
	}

	switch rune(boolByte[0]) {
	case 't':
		booleanVal = true
	case 'f':
		booleanVal = false
	default:
		return false, fmt.Errorf("Unknown boolean value (%s)", string(boolByte[0]))
	}

	var endBytes [2]byte
	_, err = io.ReadFull(req, endBytes[:])
	if err != nil {
		return false, err
	}

	if rune(endBytes[0]) != '\r' || rune(endBytes[1]) != '\n' {
		return false, errors.New("Null did not end with CRLF")
	}

	if rune(endBytes[0]) != '\r' || rune(endBytes[1]) != '\n' {
		return false, errors.New("Null did not end with CRLF")
	}

	return booleanVal, nil
}

func readNull(req io.Reader) (interface{}, error) {
	var endBytes [2]byte
	_, err := io.ReadFull(req, endBytes[:])
	if err != nil {
		return nil, err
	}

	if rune(endBytes[0]) != '\r' || rune(endBytes[1]) != '\n' {
		return nil, errors.New("Null did not end with CRLF")
	}

	if rune(endBytes[0]) != '\r' || rune(endBytes[1]) != '\n' {
		return nil, errors.New("Null did not end with CRLF")
	}

	return nil, nil
}

func readInteger(req io.Reader, isSigned bool) (int64, error) {
	var val int64
	var currByte [1]byte
	var bytesRead int
	for {
		_, err := io.ReadFull(req, currByte[:])
		if err != nil {
			return 0, err
		}

		currRune := rune(currByte[0])
		switch {
		case isSigned && currRune == '+':
			if bytesRead > 0 {
				return 0, errors.New("Number contained '+' outside of prefix")
			}
		case isSigned && currRune == '-':
			if bytesRead == 0 {
				val = -val
			} else {
				return 0, errors.New("Number contained '+' outside of prefix")
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
			_, err := io.ReadFull(req, currByte[:])
			if err != nil {
				return 0, err
			}

			if rune(currByte[0]) == '\n' {
				return val, nil
			} else {
				return 0, errors.New("Number did not end in \r\n")
			}

		default:
			return 0, fmt.Errorf("Unexpected character in number: '%s'", string(rune(currByte[0])))
		}

		bytesRead++
	}
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

type BigNumber struct {
	val      string
	negative bool
}

func readBigNumber(req io.Reader) (BigNumber, error) {
	var strVal strings.Builder
	var currByte [1]byte
	negative := false
	seenSign := false

	for {
		_, err := io.ReadFull(req, currByte[:])
		if err != nil {
			return BigNumber{}, err
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
			_, err := io.ReadFull(req, currByte[:])
			if err != nil {
				return BigNumber{}, err
			}

			if rune(currByte[0]) == '\n' {
				return BigNumber{val: strVal.String(), negative: negative}, nil
			} else {
				return BigNumber{}, errors.New("Simple String contained \r without \n")
			}
		default:
			return BigNumber{}, fmt.Errorf("Unexpected character in big number: %s", string(currByte[0]))
		}
	}
}

func readDouble(req io.Reader, isSigned bool) (float64, error) {
	var integralSign int64 = 1
	var integral int64
	var fractional int64
	var exponentSign int64 = 1
	var exponent int64

	var readState DoubleReadState
	if isSigned {
		readState = ReadIntegralSign
	} else {
		readState = ReadIntegral
	}

	var currByte [1]byte
	parsingDouble := true
	for parsingDouble {
		_, err := io.ReadFull(req, currByte[:])
		if err != nil {
			return 0, err
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
			_, err := io.ReadFull(req, nfBytes[:])
			if err != nil {
				return 0, err
			}

			if rune(nfBytes[0]) == 'n' && rune(nfBytes[1]) == 'f' {
				return math.Inf(int(integralSign)), nil
			} else {
				return 0, fmt.Errorf("Unexpected 'i%s' in double", string(nfBytes[:]))
			}

		case readState == ReadIntegral && *val == 0 && currRune == 'n':
			var anBytes [2]byte
			_, err := io.ReadFull(req, anBytes[:])
			if err != nil {
				return 0, err
			}

			if rune(anBytes[0]) == 'a' && rune(anBytes[1]) == 'n' {
				return math.NaN(), nil
			} else {
				return 0, fmt.Errorf("Unexpected 'n%s' in double", string(anBytes[:]))
			}

		case currRune == '+':
			if readState == ReadIntegralSign {
				*val = 1
				readState = ReadIntegral
			} else if readState == ReadExponentSign {
				*val = 1
				readState = ReadExponent
			} else {
				return 0, fmt.Errorf("Invalid + position (%s)", doubleReadStateStr(readState))
			}
		case currRune == '-':
			if readState == ReadIntegralSign {
				*val = -1
				readState = ReadIntegral
			} else if readState == ReadExponentSign {
				*val = -1
				readState = ReadExponent
			} else {
				return 0, fmt.Errorf("Invalid - position (%s)", doubleReadStateStr(readState))
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
			_, err := io.ReadFull(req, currByte[:])
			if err != nil {
				return 0, err
			}

			if rune(currByte[0]) == '\n' {
				parsingDouble = false
			} else {
				return 0, errors.New("Number did not end in \r\n")
			}

		default:
			return 0, fmt.Errorf("Unexpected character in double (%s): %v", doubleReadStateStr(readState), rune(currByte[0]))
		}
	}

	var doubleVal float64

	doubleVal = float64(integral * integralSign)

	fDigits := float64(len(fmt.Sprint(fractional)))
	doubleVal = doubleVal + (float64(fractional) / math.Pow(10, fDigits))
	doubleVal = doubleVal * math.Pow(10, float64(exponent*exponentSign))
	return doubleVal, nil
}

type VerbatimString struct {
	encoding string
	data     []byte
}

func readVerbatimString(req io.Reader) (VerbatimString, error) {
	dataLength, err := readInteger(req, false)
	if err != nil {
		return VerbatimString{}, err
	}

	var encodingBytes [3]byte
	_, err = io.ReadFull(req, encodingBytes[:])
	if err != nil {
		return VerbatimString{}, err
	}

	var colonByte [1]byte
	_, err = io.ReadFull(req, colonByte[:])
	if err != nil {
		return VerbatimString{}, err
	}

	if string(colonByte[:]) != ":" {
		return VerbatimString{}, fmt.Errorf("Expected colon but got %s", string(colonByte[:]))
	}

	dataBytes := make([]byte, dataLength)
	_, err = io.ReadFull(req, dataBytes)
	if err != nil {
		return VerbatimString{}, err
	}

	var endBytes [2]byte
	_, err = io.ReadFull(req, endBytes[:])
	if err != nil {
		return VerbatimString{}, err
	}

	if rune(endBytes[0]) != '\r' || rune(endBytes[1]) != '\n' {
		return VerbatimString{}, errors.New("Verbatim String did not end with CRLF")
	}

	return VerbatimString{encoding: string(encodingBytes[:]), data: dataBytes}, nil
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
