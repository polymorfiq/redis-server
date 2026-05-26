package commands

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type Command interface {
	Parse([]string) error
	Definition() CommandDefinition
	Execute(*client.Session) error
}

var allCommands = map[string]func() Command{
	"set":    NewSet,
	"get":    NewGet,
	"ping":   NewPing,
	"hello":  NewHello,
	"echo":   NewEcho,
	"lpush":  NewLPush,
	"rpush":  NewRPush,
	"llen":   NewLLen,
	"lrange": NewLRange,
}

func ParseCommand(cmdArray *resp.Array) (Command, error) {
	if len(cmdArray.Values) == 0 {
		return nil, fmt.Errorf("no command specified")
	}

	cmdStrs := make([]string, 0, len(cmdArray.Values))
	for _, cmdVal := range cmdArray.Values {
		valStr, isString := cmdVal.(fmt.Stringer)

		if !isString {
			return nil, fmt.Errorf("expected command string and got a %T (%v)", cmdVal, cmdVal)
		}

		cmdStrs = append(cmdStrs, valStr.String())
	}

	cmdStr := strings.ToLower(cmdStrs[0])
	cmd, exists := allCommands[cmdStr]

	if !exists && cmdStr == "command" {
		return NewListCommands(), nil
	}
	if !exists {
		return nil, fmt.Errorf("Unknown Request: %s\n", cmdStrs[0])
	}

	cmdInst := cmd()
	err := cmdInst.Parse(cmdStrs[1:])
	if err != nil {
		return nil, err
	}

	return cmdInst, nil
}

type OptionalValue[T any] struct {
	Active bool
	Value  T
}

func ActiveOptionalValue[T any](val T) OptionalValue[T] {
	opt := OptionalValue[T]{}
	opt.Active = true
	opt.Value = val
	return opt
}
