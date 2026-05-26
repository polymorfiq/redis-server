package main

import (
	"fmt"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/commands"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/storage_engine"
)

var _ = net.Listen
var _ = os.Exit

func main() {
	l, err := net.Listen("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	storage := storage_engine.New()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		c := client.New(conn)
		session := client.NewSession(c, storage)

		go handleConnection(session)
	}
}

func handleConnection(session *client.Session) {
	fmt.Printf("New connection from %v\n", session.RemoteAddr())

	for {
		val, err := session.ReadNext()
		if err != nil {
			_ = session.LogError(fmt.Sprintf("Error reading from session: %s", err))
			break
		}

		cmdArray, isArray := val.(*resp.Array)
		if !isArray {
			_ = session.LogError(fmt.Sprintf("non-array command received (Got %T): %v", val, val))
			break
		}

		cmd, err := commands.ParseCommand(cmdArray)
		if err != nil {
			_ = session.LogError(fmt.Sprintf("error parsing command (%v): %v", cmdArray, err))
			continue
		}

		err = cmd.Execute(session)
		if err != nil {
			_ = session.LogError(fmt.Sprintf("error executing command (%v): %v", cmdArray, err))
			continue
		}
	}
}
