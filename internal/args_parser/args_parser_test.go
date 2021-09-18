package args_parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"ponglehub.co.uk/tools/mudly/internal/args_parser"
	"ponglehub.co.uk/tools/mudly/internal/target"
)

func TestParse(t *testing.T) {
	for _, test := range []struct {
		Name            string
		Args            []string
		ExpectedCommand args_parser.CommandType
		ExpectedOptions *args_parser.Options
		ExpectedError   string
	}{
		{
			Name:            "bare artefact",
			Args:            []string{"+image"},
			ExpectedCommand: args_parser.BUILD_COMMAND,
			ExpectedOptions: &args_parser.Options{Targets: []target.Target{
				{Dir: ".", Artefact: "image"},
			}},
		},
	} {
		t.Run(test.Name, func(u *testing.T) {
			command, options, err := args_parser.Parse(test.Args)

			assert.Equal(u, test.ExpectedCommand, command)

			if test.ExpectedOptions != nil {
				assert.Equal(u, *test.ExpectedOptions, options)
			}

			if test.ExpectedError != "" {
				assert.EqualError(u, err, test.ExpectedError)
			}
		})
	}
}
