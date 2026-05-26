package commands

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
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

	poppedVals := make([]resp.Value, 0, cmd.NumPop)
	for range cmd.NumPop {
		currArray, popped, err := storage.LPop(cmd.Key)
		if err != nil {
			return err
		}

		if currArray == nil || popped == nil {
			continue
		}

		if !cmd.ArrayPop {
			return sess.Send(popped)
		}

		poppedVals = append(poppedVals, popped)
	}

	if len(poppedVals) == 0 && sess.IsRESP3() {
		return sess.Send(resp.NewNull())
	} else if len(poppedVals) == 0 {
		return sess.Send(resp.NullBulkString())
	}

	popArray := resp.NewArray().(*resp.Array)
	popArray.Values = poppedVals
	return sess.Send(popArray)
}
