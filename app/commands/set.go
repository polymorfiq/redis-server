package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/storage_engine"
)

type Set struct {
	Key     string
	Value   resp.Value
	NX      OptionalValue[bool]
	XX      OptionalValue[bool]
	IFEQ    OptionalValue[string]
	IFNE    OptionalValue[string]
	IFDEQ   OptionalValue[string]
	IFDNE   OptionalValue[string]
	GET     OptionalValue[bool]
	EX      OptionalValue[uint64]
	PX      OptionalValue[uint64]
	EXAT    OptionalValue[uint64]
	PXAT    OptionalValue[uint64]
	KEEPTTL OptionalValue[uint64]
}

func NewSet() Command {
	return &Set{}
}

func (cmd *Set) Definition() CommandDefinition {
	return NewCommandDefinition("SET", 2)
}

func (cmd *Set) Parse(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("SET expects at least two arguments")
	}

	cmd.Key = args[0]
	cmd.Value = resp.MaybeSimpleString(args[1])
	args = args[2:]

	idx := 0
	for idx < len(args) {
		isLastArg := idx == len(args)-1

		optName := strings.ToLower(args[idx])
		switch optName {
		case "ex":
			if isLastArg {
				return errors.New("option EX expects argument")
			}

			exSeconds, err := strconv.Atoi(args[idx+1])
			if err != nil {
				return err
			}
			cmd.EX = ActiveOptionalValue(uint64(exSeconds))
			idx += 2

		case "px":
			if isLastArg {
				return errors.New("option PX expects argument")
			}

			pxMilli, err := strconv.Atoi(args[idx+1])
			if err != nil {
				return err
			}
			cmd.PX = ActiveOptionalValue(uint64(pxMilli))
			idx += 2

		case "exat":
			if isLastArg {
				return errors.New("option EXAT expects argument")
			}

			expiresAt, err := strconv.Atoi(args[idx+1])
			if err != nil {
				return err
			}
			cmd.EXAT = ActiveOptionalValue(uint64(expiresAt))
			idx += 2

		default:
			return fmt.Errorf("unkown SET option %s", args[idx])
		}
	}

	return cmd.UnpackOptions(args[2:])
}

func (cmd *Set) Execute(sess *client.Session) error {
	storage := sess.Storage()

	storageOpts := storage_engine.StorageOpts{}
	if cmd.EX.Active {
		storageOpts.SetExpiresAt(time.Now().Add(time.Duration(cmd.EX.Value) * time.Second))
	}
	if cmd.PX.Active {
		storageOpts.SetExpiresAt(time.Now().Add(time.Duration(cmd.PX.Value) * time.Millisecond))
	}

	if cmd.EXAT.Active {
		storageOpts.SetExpiresAt(time.Unix(int64(cmd.EXAT.Value), 0))
	}

	err := storage.Put(cmd.Key, cmd.Value, storageOpts)

	if err != nil {
		return err
	}

	return sess.Send(resp.SimpleStringFromString("OK"))
}

func (cmd *Set) UnpackOptions(args []string) error {
	if len(args) == 0 {
		return nil
	}

	switch strings.ToLower(args[0]) {
	case "nx":
		cmd.NX = OptionalValue[bool]{true, true}
		args = args[1:]
	case "xx":
		cmd.XX = OptionalValue[bool]{true, true}
	default:
		return fmt.Errorf("unknown option: %s", args[0])
	}

	return cmd.UnpackOptions(args)
}
