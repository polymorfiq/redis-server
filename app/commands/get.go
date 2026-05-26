package commands

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type Get struct {
	Key string
}

func NewGet() Command {
	return &Get{}
}

func (cmd *Get) Definition() CommandDefinition {
	return NewCommandDefinition("GET", 1)
}

func (cmd *Get) Parse(args []string) error {
	cmd.Key = args[0]
	return cmd.UnpackOptions(args[1:])
}

func (cmd *Get) Execute(sess *client.Session) error {
	storage := sess.Storage()
	val, exists := storage.Get(cmd.Key)
	if !exists && sess.IsRESP3() {
		return sess.Send(resp.NewNull())
	} else if !exists {
		return sess.Send(resp.NullBulkString())
	}

	return sess.Send(val)
}

func (cmd *Get) UnpackOptions(args []string) error {
	if len(args) == 0 {
		return nil
	}

	switch strings.ToLower(args[0]) {
	default:
		return fmt.Errorf("unknown option: %s", args[0])
	}
}
