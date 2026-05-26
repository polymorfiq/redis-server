package commands

import (
	"context"
	"errors"
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

	popResp := resp.NewArray().(*resp.Array)
	popResp.Values = append(popResp.Values, resp.BulkStringFromString(cmd.Key))
	for {
		_, popped, err := storage.LPop(cmd.Key)
		if err != nil && !errors.Is(err, storage_engine.NotArrayError) {
			return err
		}

		if popped == nil {
			ctx := context.Background()
			if cmd.HasTimeout {
				ctx, _ = context.WithTimeout(ctx, cmd.Timeout)
			}

			select {
			case <-storage.ChangeChannel(cmd.Key):
				continue

			case <-ctx.Done():
				return sess.Send(resp.NullArray())
			}
		} else {
			popResp.Values = append(popResp.Values, popped)
			return sess.Send(popResp)
		}
	}
}
