package commands

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type LRange struct {
	Key        string
	StartIndex int
	StopIndex  int
}

func NewLRange() Command {
	return &LRange{}
}

func (cmd *LRange) Definition() CommandDefinition {
	return NewCommandDefinition("LRANGE", 2)
}

func (cmd *LRange) Parse(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("LRANGE expects at least 3 arguments")
	}

	cmd.Key = args[0]
	startIdx, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}

	stopIdx, err := strconv.Atoi(args[2])
	if err != nil {
		return err
	}

	cmd.StartIndex = startIdx
	cmd.StopIndex = stopIdx

	return nil
}

func (cmd *LRange) Execute(sess *client.Session) error {
	storage := sess.Storage()
	curr, exists := storage.Get(cmd.Key)
	if !exists {
		curr = resp.NewArray()
	}

	currArray, isArray := curr.(*resp.Array)
	if !isArray {
		return fmt.Errorf("%s is not array (%T)", cmd.Key, curr)
	}

	if cmd.StartIndex < 0 {
		cmd.StartIndex = max(len(currArray.Values)-cmd.StartIndex, 0)
	}

	if cmd.StopIndex < 0 {
		cmd.StopIndex = max(len(currArray.Values)-cmd.StopIndex, 0)
	}

	if cmd.StartIndex > len(currArray.Values) || cmd.StartIndex > cmd.StopIndex {
		return sess.Send(resp.NewArray())
	}

	if cmd.StopIndex >= len(currArray.Values) {
		cmd.StopIndex = len(currArray.Values) - 1
	}

	rangeArray := resp.NewArray().(*resp.Array)
	for i := cmd.StartIndex; i <= cmd.StopIndex; i++ {
		rangeArray.Values = append(rangeArray.Values, currArray.Values[i])
	}

	return sess.Send(rangeArray)
}
