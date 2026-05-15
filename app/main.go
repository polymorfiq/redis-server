package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

var _ = net.Listen
var _ = os.Exit

func main() {
	l, err := net.Listen("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	for {
		line, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Fatalf("Error reading line: %s", err.Error())
		}

		_, err = conn.Write([]byte("+PONG\r\n"))
		if err != nil {
			log.Fatalf("Error writing line: %s", err.Error())
		}

		fmt.Println(line)
	}
}
