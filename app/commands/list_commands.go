package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type ListCommands struct {
	Definitions []CommandDefinition
}

func NewListCommands() Command {
	var defs []CommandDefinition
	for _, newCmd := range allCommands {
		defs = append(defs, newCmd().Definition())
	}

	return &ListCommands{
		Definitions: defs,
	}
}

func (cmd *ListCommands) Parse(args []string) error {
	if len(args) == 0 {
		return nil
	}

	return fmt.Errorf("unknown command %s", args[0])
}

func (cmd *ListCommands) Definition() CommandDefinition {
	return NewCommandDefinition("COMMAND", 0)
}

func (cmd *ListCommands) Execute(sess *client.Session) error {
	cmdList := resp.NewArray().(*resp.Array)
	cmdList.Values = make([]resp.Value, 0, len(cmd.Definitions))

	for _, cmdDef := range cmd.Definitions {
		cmdList.Values = append(cmdList.Values, cmdDef.ToRespArray())
	}

	return sess.Send(cmdList)
}

type CommandDefinition struct {
	Name          string
	Arity         int64
	Flags         []string
	FirstKey      int64
	LastKey       int64
	Step          int64
	AclCategories []string
	CommandTips   []string
	KeySpecs      *resp.Map
	Subcommands   []CommandDefinition
}

func NewCommandDefinition(name string, arity int) CommandDefinition {
	return CommandDefinition{
		Name:  name,
		Arity: int64(arity),
	}
}

func (def *CommandDefinition) ToRespArray() *resp.Array {
	subcommands := resp.NewArray().(*resp.Array)
	subcommands.Values = make([]resp.Value, 0, len(def.Subcommands))
	for _, cmd := range def.Subcommands {
		subcommands.Values = append(subcommands.Values, cmd.ToRespArray())
	}

	return resp.ArrayFromValues([]resp.Value{
		resp.SimpleStringFromString(def.Name),
		resp.IntegerFromInt(def.Arity),
		resp.ArrayOfStrings(def.Flags),
		resp.IntegerFromInt(def.FirstKey),
		resp.IntegerFromInt(def.LastKey),
		resp.IntegerFromInt(def.Step),
		resp.ArrayOfStrings(def.AclCategories),
		resp.ArrayOfStrings(def.CommandTips),
		def.KeySpecs,
		subcommands,
	})
}

type KeySpec struct {
	Type  string
	Index int
	Spec  string
}
