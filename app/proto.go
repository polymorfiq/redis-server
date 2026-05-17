package main

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

var once sync.Once
var saved map[string]string

func handleReq(req interface{}, resp io.Writer) error {
	once.Do(func() {
		saved = make(map[string]string)
	})

	reqInterfaceArray, _ := req.([]interface{})
	reqArray := make([]string, len(reqInterfaceArray))
	for i, v := range reqInterfaceArray {
		reqArray[i] = v.(string)
	}

	cmd := strings.ToLower(reqArray[0])
	args := reqArray[1:]
	switch {
	case cmd == "ping":
		if _, err := io.WriteString(resp, "+PONG\r\n"); err != nil {
			return err
		}

	case cmd == "echo":
		if err := writeBulkString(resp, strings.Join(args, " ")); err != nil {
			return err
		}

	case cmd == "set":
		saved[args[0]] = args[1]

		if err := writeSimpleString(resp, "OK"); err != nil {
			return err
		}

	case cmd == "get":
		if err := writeBulkString(resp, saved[args[0]]); err != nil {
			return err
		}

	case cmd == "hello":
		fmt.Println("HELLO REQ\n")

	default:
		fmt.Printf("Unknown Request: %v\n", req)
	}

	return nil
}
