package commands

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/storage_engine"
)

type LPop struct {
	Key      string
	NumPop   uint64
	ArrayPop bool
}

func NewLPop() Command {
	return &LPop{NumPop: 1, ArrayPop: false}
}

func (cmd *LPop) Definition() CommandDefinition {
	return NewCommandDefinition("LPOP", 2)
}

func (cmd *LPop) Parse(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("LPOP expects at least 1 arguments")
	}

	cmd.Key = args[0]

	if len(args) > 1 {
		n, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("LPOP expects numeric argument")
		}

		cmd.NumPop = n
		cmd.ArrayPop = true
	}

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

	var newVal *resp.Array
	var popped resp.Value
	if cmd.ArrayPop {
		newVal, popped = cmd.bulkPop(currArray, cmd.NumPop)
	} else {
		newVal, popped = cmd.singlePop(currArray)
	}

	err := storage.Put(cmd.Key, newVal, storage_engine.StorageOpts{})
	if err != nil {
		return err
	}

	return sess.Send(popped)
}

func (cmd *LPop) singlePop(currArray *resp.Array) (*resp.Array, resp.Value) {
	if len(currArray.Values) == 0 {
		return currArray, resp.NullBulkString()
	}

	newArray := resp.NewArray().(*resp.Array)
	popped := currArray.Values[0]
	if len(currArray.Values) > 1 {
		newArray.Values = currArray.Values[1:]
	} else {
		newArray.Values = nil
	}

	return newArray, popped
}

func (cmd *LPop) bulkPop(currArray *resp.Array, n uint64) (newVal *resp.Array, popped *resp.Array) {
	if len(currArray.Values) == 0 {
		return currArray, resp.NewArray().(*resp.Array)
	}

	if len(currArray.Values) < int(n) {
		return resp.NewArray().(*resp.Array), currArray
	}

	popped = resp.NewArray().(*resp.Array)
	popped.Values = make([]resp.Value, 0, n)
	for i := uint64(0); i < n; i++ {
		popped.Values = append(popped.Values, currArray.Values[i])
	}

	afterPop := resp.NewArray().(*resp.Array)
	afterPop.Values = currArray.Values[n:]

	return afterPop, popped
}
