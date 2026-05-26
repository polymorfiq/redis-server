package commands

import (
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type Echo struct {
	Content string
}

func NewEcho() Command {
	return &Echo{}
}

func (cmd *Echo) Parse(args []string) error {
	cmd.Content = strings.Join(args, " ")
	return nil
}

func (cmd *Echo) Definition() CommandDefinition {
	return NewCommandDefinition("ECHO", 1)
}

func (cmd *Echo) Execute(sess *client.Session) error {
	return sess.Send(resp.BulkStringFromString(cmd.Content))
}
