package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type Ping struct{}

func NewPing() Command {
	return &Ping{}
}

func (cmd *Ping) Definition() CommandDefinition {
	return NewCommandDefinition("PING", 0)
}

func (cmd *Ping) Parse(args []string) error {
	if len(args) == 0 {
		return nil
	}

	return fmt.Errorf("unknown command %s", args[0])
}

var pong = resp.SimpleStringFromString("PONG")

func (cmd *Ping) Execute(sess *client.Session) error {
	return sess.Send(pong)
}
