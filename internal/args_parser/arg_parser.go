package args_parser

import (
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

func Parse(args []string) (CommandType, Options, error) {
	options := Options{}

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
		}
	}

	return BUILD_COMMAND,
		options,
		nil
}
