package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/storage_engine"
)

type RPush struct {
	Key    string
	Values []resp.Value
}

func NewRPush() Command {
	return &RPush{}
}

func (cmd *RPush) Definition() CommandDefinition {
	return NewCommandDefinition("RPUSH", 2)
}

func (cmd *RPush) Parse(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("RPUSH expects at least 2 arguments")
	}

	cmd.Key = args[0]
	cmd.Values = make([]resp.Value, 0, len(args[1:]))
	for i := 1; i < len(args); i++ {
		cmd.Values = append(cmd.Values, resp.BulkStringFromString(args[i]))
	}

	return nil
}

func (cmd *RPush) Execute(sess *client.Session) error {
	storage := sess.Storage()
	curr, exists := storage.Get(cmd.Key)
	if !exists {
		curr = resp.NewArray()
	}

	currArray, isArray := curr.(*resp.Array)
	if !isArray {
		return fmt.Errorf("%s is not array (%T)", cmd.Key, curr)
	}
	for _, val := range cmd.Values {
		currArray.Values = append(currArray.Values, val)
	}

	err := storage.Put(cmd.Key, currArray, storage_engine.StorageOpts{})
	if err != nil {
		return err
	}

	return sess.Send(resp.IntegerFromInt(int64(len(currArray.Values))))
}
