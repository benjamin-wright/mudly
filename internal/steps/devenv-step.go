package steps

import (
	"fmt"

	"ponglehub.co.uk/tools/mudly/internal/runner"
)

type DevenvStep struct {
	Name    string
	Compose string
}

func (d DevenvStep) isRunning(dir string, artefact string) bool {
	return !runShellCommand(&shellCommand{
		dir:      dir,
		artefact: artefact,
		step:     fmt.Sprintf("%s (check)", d.Name),
		command:  "bash",
		args: []string{
			"-c",
			fmt.Sprintf("docker compose ls | grep \"%s\"", d.Name),
		},
		stdin: d.Compose,
	})
}

func (d DevenvStep) Run(dir string, artefact string, env map[string]string) runner.CommandResult {
	if d.isRunning(dir, artefact) {
		return runner.COMMAND_SKIPPED
	}

	success := runShellCommand(&shellCommand{
		dir:      dir,
		artefact: artefact,
		step:     d.Name,
		command:  "docker",
		args: []string{
			"compose",
			"--project-name",
			d.Name,
			"-f",
			"-",
			"up",
			"-d",
		},
		stdin: d.Compose,
	})

	if success {
		return runner.COMMAND_SUCCESS
	} else {
		return runner.COMMAND_ERROR
	}
}

func (d DevenvStep) String() string {
	return d.Name
}
