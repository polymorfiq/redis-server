package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type Hello struct {
	ProtoVersion int
}

func NewHello() Command {
	return &Hello{
		ProtoVersion: 3,
	}
}

func (cmd *Hello) Definition() CommandDefinition {
	return NewCommandDefinition("HELLO", 0)
}

func (cmd *Hello) Parse(args []string) error {
	if len(args) == 0 {
		return nil
	}

	return fmt.Errorf("unknown options %s", args)
}

func (cmd *Hello) Execute(sess *client.Session) error {
	sess.SetProtoVersion(cmd.ProtoVersion)

	return sess.Send(resp.SimpleStringFromString("OK"))
}
