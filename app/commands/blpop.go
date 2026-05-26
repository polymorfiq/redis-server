package commands

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/storage_engine"
)

type BLPop struct {
	Key        string
	Timeout    time.Duration
	HasTimeout bool
}

func NewBLPop() Command {
	return &BLPop{}
}

func (cmd *BLPop) Definition() CommandDefinition {
	return NewCommandDefinition("BLPOP", 2)
}

func (cmd *BLPop) Parse(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("BLPOP expects at least 1 arguments")
	}

	cmd.Key = args[0]
	if len(args) == 2 {
		timeout, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("BLPOP expects numeric argument")
		}

		cmd.HasTimeout = timeout > 0
		cmd.Timeout = time.Duration(timeout) * time.Second
	}

	return nil
}

func (cmd *BLPop) Execute(sess *client.Session) error {
	storage := sess.Storage()

	var newVal *resp.Array
	var popped resp.Value
	keepLooping := true
	for keepLooping {
		curr, exists := storage.Get(cmd.Key)
		if exists {
			currArray, isArray := curr.(*resp.Array)
			if isArray && len(currArray.Values) >= 1 {
				newVal, popped = cmd.singlePop(currArray)
				break
			}
		}

		ctx := context.Background()
		if cmd.HasTimeout {
			ctx, _ = context.WithTimeout(ctx, cmd.Timeout)
		}

		select {
		case <-storage.ChangeChannel(cmd.Key):
			continue

		case <-ctx.Done():
			newVal = nil
			popped = resp.NullArray()
			keepLooping = false
		}

	}

	if newVal != nil {
		err := storage.Put(cmd.Key, newVal, storage_engine.StorageOpts{})
		if err != nil {
			return err
		}
	}

	return sess.Send(popped)
}

func (cmd *BLPop) singlePop(currArray *resp.Array) (*resp.Array, resp.Value) {
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
