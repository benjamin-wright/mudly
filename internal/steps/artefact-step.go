package steps

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"ponglehub.co.uk/tools/mudly/internal/runner"
)

type ArtefactStep struct {
	Condition string
}

func (a ArtefactStep) Run(dir string, artefact string, env map[string]string) runner.CommandResult {
	if a.Condition != "" {
		test := runShellCommand(&shellCommand{
			dir:      dir,
			artefact: artefact,
			step:     fmt.Sprintf("%s (test)", artefact),
			command:  "/bin/bash",
			args:     []string{"-c", a.Condition},
			env:      env,
			test:     true,
		})

		if !test {
			logrus.Debugf("%s[(test)]: Skipping artefact", artefact)
			return runner.COMMAND_SKIP_ARTEFACT
		}
	}

	return runner.COMMAND_SUCCESS
}

func (a ArtefactStep) String() string {
	return "cache"
}
