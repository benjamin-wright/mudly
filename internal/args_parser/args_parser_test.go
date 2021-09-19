package args_parser_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"ponglehub.co.uk/tools/mudly/internal/args_parser"
	"ponglehub.co.uk/tools/mudly/internal/target"
)

func TestParse(t *testing.T) {
	for _, test := range []struct {
		Name     string
		Args     []string
		Expected []target.Target
	}{
		{
			Name:     "bare artefact",
			Args:     []string{"+image"},
			Expected: []target.Target{{Dir: ".", Artefact: "image"}},
		},
		{
			Name:     "relative artefact",
			Args:     []string{"../sibling+image"},
			Expected: []target.Target{{Dir: "../sibling", Artefact: "image"}},
		},
		{
			Name: "multiple artefacts",
			Args: []string{"+image", "../sibling+image"},
			Expected: []target.Target{
				{Dir: ".", Artefact: "image"},
				{Dir: "../sibling", Artefact: "image"},
			},
		},
	} {
		t.Run(fmt.Sprintf("Target parsing: %s", test.Name), func(u *testing.T) {
			command, options, err := args_parser.Parse(test.Args)

			assert.Equal(u, args_parser.BUILD_COMMAND, command)

			if test.Expected != nil {
				assert.Equal(u, args_parser.Options{Targets: test.Expected}, options)
			}

			assert.Nil(u, err)
		})
	}

	for _, test := range []struct {
		Name             string
		Args             []string
		ExpectedNoDeps   bool
		ExpectedOnlyDeps bool
	}{
		{Name: "no flags", Args: []string{"+image"}},
		{Name: "nodeps short", Args: []string{"--no-deps", "+image"}, ExpectedNoDeps: true},
		{Name: "nodeps long", Args: []string{"--no-dependencies", "+image"}, ExpectedNoDeps: true},
		{Name: "nodeps short", Args: []string{"--deps", "+image"}, ExpectedOnlyDeps: true},
		{Name: "nodeps long", Args: []string{"--dependencies", "+image"}, ExpectedOnlyDeps: true},
	} {
		t.Run(fmt.Sprintf("Target flags: %s", test.Name), func(u *testing.T) {
			command, options, err := args_parser.Parse(test.Args)

			assert.Equal(u, args_parser.BUILD_COMMAND, command)

			assert.Equal(
				u,
				args_parser.Options{
					NoDeps:   test.ExpectedNoDeps,
					OnlyDeps: test.ExpectedOnlyDeps,
					Targets:  []target.Target{{Dir: ".", Artefact: "image"}},
				},
				options,
			)

			assert.Nil(u, err)
		})
	}

	t.Run("Stop command", func(u *testing.T) {
		command, options, err := args_parser.Parse([]string{"stop"})

		assert.Equal(u, args_parser.STOP_COMMAND, command)
		assert.Equal(u, args_parser.Options{}, options)
		assert.Nil(u, err)
	})

	for _, test := range []struct {
		Name  string
		Args  []string
		Error string
	}{
		{Name: "no args", Args: []string{}, Error: "failed parsing args: no args provided"},
		{Name: "deps and no deps form1", Args: []string{"--no-deps", "--deps", "+image"}, Error: "dependecies and no-dependencies options cannot be used together"},
		{Name: "deps and no deps form2", Args: []string{"--no-deps", "--dependencies", "+image"}, Error: "dependecies and no-dependencies options cannot be used together"},
		{Name: "deps and no deps form3", Args: []string{"--no-dependencies", "--deps", "+image"}, Error: "dependecies and no-dependencies options cannot be used together"},
		{Name: "deps and no deps form4", Args: []string{"--no-dependencies", "--dependencies", "+image"}, Error: "dependecies and no-dependencies options cannot be used together"},
		{Name: "only deps flag form1", Args: []string{"--deps"}, Error: "must provide a target"},
		{Name: "only deps flag form2", Args: []string{"--dependencies"}, Error: "must provide a target"},
		{Name: "no deps flag form1", Args: []string{"--no-deps"}, Error: "must provide a target"},
		{Name: "no deps flag form2", Args: []string{"--no-dependencies"}, Error: "must provide a target"},
		{Name: "stop with flags 1", Args: []string{"--deps", "stop"}, Error: "stop command does not accept flags or arguments"},
		{Name: "stop with flags 2", Args: []string{"--no-deps", "stop"}, Error: "stop command does not accept flags or arguments"},
		{Name: "stop with flags 3", Args: []string{"--dependencies", "stop"}, Error: "stop command does not accept flags or arguments"},
		{Name: "stop with flags 4", Args: []string{"--no-dependencies", "stop"}, Error: "stop command does not accept flags or arguments"},
		{Name: "stop with targets", Args: []string{"stop", "+image"}, Error: "stop command does not accept flags or arguments"},
		{Name: "stop in build", Args: []string{"--no-deps", "+image", "stop"}, Error: "stop command does not accept flags or arguments"},
	} {
		t.Run(fmt.Sprintf("Errors: %s", test.Name), func(u *testing.T) {
			command, options, err := args_parser.Parse(test.Args)

			assert.Equal(u, args_parser.NO_COMMAND, command)
			assert.Equal(u, args_parser.Options{}, options)
			assert.EqualError(u, err, test.Error)
		})
	}
}
