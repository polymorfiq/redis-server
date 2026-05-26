package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type Hello struct{}

func NewHello() Command {
	return &Hello{}
}

func (cmd *Hello) Definition() CommandDefinition {
	return NewCommandDefinition("HELLO", 0)
}

func (cmd *Hello) Parse(args []string) error {
	if len(args) == 0 {
		return nil
	}

	return fmt.Errorf("unknown command %s", args[0])
}

func (cmd *Hello) Execute(sess *client.Session) error {
	return sess.Send(resp.SimpleStringFromString("OK"))
}
