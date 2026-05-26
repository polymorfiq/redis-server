package commands

import (
	"fmt"
	"slices"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/storage_engine"
)

type LPush struct {
	Key    string
	Values []resp.Value
}

func NewLPush() Command {
	return &LPush{}
}

func (cmd *LPush) Definition() CommandDefinition {
	return NewCommandDefinition("LPUSH", 2)
}

func (cmd *LPush) Parse(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("LPUSH expects at least 2 arguments")
	}

	cmd.Key = args[0]
	cmd.Values = make([]resp.Value, 0, len(args[1:]))
	for i := 1; i < len(args); i++ {
		cmd.Values = append(cmd.Values, resp.BulkStringFromString(args[i]))
	}

	return nil
}

func (cmd *LPush) Execute(sess *client.Session) error {
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
		currArray.Values = slices.Concat([]resp.Value{val}, currArray.Values)
	}

	err := storage.Put(cmd.Key, currArray, storage_engine.StorageOpts{})
	if err != nil {
		return err
	}

	return sess.Send(resp.IntegerFromInt(int64(len(currArray.Values))))
}
