package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/storage_engine"
)

type LPop struct {
	Key string
}

func NewLPop() Command {
	return &LPop{}
}

func (cmd *LPop) Definition() CommandDefinition {
	return NewCommandDefinition("LPOP", 2)
}

func (cmd *LPop) Parse(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("LPOP expects at least 1 arguments")
	}

	cmd.Key = args[0]

	return nil
}

func (cmd *LPop) Execute(sess *client.Session) error {
	storage := sess.Storage()
	curr, exists := storage.Get(cmd.Key)
	if !exists {
		return sess.Send(resp.NullBulkString())
	}

	currArray, isArray := curr.(*resp.Array)
	if !isArray {
		return fmt.Errorf("%s is not array (%T)", cmd.Key, curr)
	}

	if len(currArray.Values) == 0 {
		return sess.Send(resp.NullBulkString())
	}

	popped := currArray.Values[0]
	if len(currArray.Values) > 1 {
		currArray.Values = currArray.Values[1:]
	} else {
		currArray.Values = nil
	}

	err := storage.Put(cmd.Key, currArray, storage_engine.StorageOpts{})
	if err != nil {
		return err
	}

	return sess.Send(popped)
}
