package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type LLen struct {
	Key string
}

func NewLLen() Command {
	return &LLen{}
}

func (cmd *LLen) Definition() CommandDefinition {
	return NewCommandDefinition("LLEN", 2)
}

func (cmd *LLen) Parse(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("LLEN expects at least 1 arguments")
	}

	cmd.Key = args[0]

	return nil
}

func (cmd *LRange) Execute(sess *client.Session) error {
	storage := sess.Storage()
	curr, exists := storage.Get(cmd.Key)
	if !exists {
		return sess.Send(resp.IntegerFromInt(0))
	}

	currArray, isArray := curr.(*resp.Array)
	if !isArray {
		return fmt.Errorf("%s is not array (%T)", cmd.Key, curr)
	}

	return sess.Send(resp.IntegerFromInt(int64(len(currArray.Values))))
}
