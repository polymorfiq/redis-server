package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

var _ = net.Listen
var _ = os.Exit

var connMut sync.Mutex
var conns []net.Conn

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
		line, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading from connection (%s): %s\n", conn.RemoteAddr(), err.Error())
		}

		line = strings.TrimSuffix(line, "\n")
		fmt.Printf("MSG (%s): %s\n", conn.RemoteAddr(), line)

		_, err = conn.Write([]byte("+PONG\r\n"))
		if err != nil {
			fmt.Printf("Error writing to connection (%s): %s\n", conn.RemoteAddr(), err.Error())
		}
	}
}
