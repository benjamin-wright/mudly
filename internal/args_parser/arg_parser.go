package args_parser

import (
	"errors"
	"fmt"
	"strings"

	"ponglehub.co.uk/tools/mudly/internal/target"
)

type CommandType int

const (
	NO_COMMAND CommandType = iota
	BUILD_COMMAND
	STOP_COMMAND
)

type Options struct {
	Targets  []target.Target
	NoDeps   bool
	OnlyDeps bool
}

func (o Options) validate() error {
	if o.NoDeps && o.OnlyDeps {
		return errors.New("dependecies and no-dependencies options cannot be used together")
	}

	emptyDeps := o.OnlyDeps && len(o.Targets) == 0
	emptyNoDeps := o.NoDeps && len(o.Targets) == 0

	if emptyDeps || emptyNoDeps {
		return errors.New("must provide a target")
	}

	return nil
}

func Parse(args []string) (CommandType, Options, error) {
	if len(args) == 0 {
		return NO_COMMAND, Options{}, errors.New("failed parsing args: no args provided")
	}

	options := Options{}
	isStopCommand := false

	for _, value := range args {
		if strings.Contains(value, "+") {
			target, err := target.ParseTarget(value)
			if err != nil {
				return NO_COMMAND, Options{}, fmt.Errorf("failed parsing args: %+v", err)
			}

			if target == nil {
				return NO_COMMAND, Options{}, fmt.Errorf("failed parsing args: null target for input '%s'", value)
			}

			options.Targets = append(options.Targets, *target)
			continue
		}

		if strings.HasPrefix(value, "--") {
			switch value {
			case "--no-deps", "--no-dependencies":
				options.NoDeps = true
				continue
			case "--deps", "--dependencies":
				options.OnlyDeps = true
				continue
			default:
				return NO_COMMAND, Options{}, fmt.Errorf("unrecognised flag: %s", value)
			}
		}

		if value == "stop" {
			isStopCommand = true
			continue
		}
	}

	if isStopCommand {
		if options.NoDeps || options.OnlyDeps || len(options.Targets) > 0 {
			return NO_COMMAND, Options{}, errors.New("stop command does not accept flags or arguments")
		}

		return STOP_COMMAND, Options{}, nil
	}

	err := options.validate()
	if err != nil {
		return NO_COMMAND, Options{}, err
	}

	return BUILD_COMMAND,
		options,
		nil
}
