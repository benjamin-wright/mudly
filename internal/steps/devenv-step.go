package steps

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"ponglehub.co.uk/tools/mudly/internal/runner"
)

type DevenvStep struct {
	Name    string
	Compose string
}

func (d DevenvStep) isRunning(dir string, artefact string) bool {
	return runShellCommand(&shellCommand{
		dir:      dir,
		artefact: artefact,
		step:     fmt.Sprintf("%s (check)", d.Name),
		command:  "/bin/bash",
		args: []string{
			"-c",
			fmt.Sprintf("docker compose ls | grep \"mudly__%s\"", d.Name),
		},
		stdin: d.Compose,
		test:  true,
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
			fmt.Sprintf("mudly__%s", d.Name),
			"-f",
			"-",
			"up",
			"-d",
			"--quiet-pull",
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

func CleanupDevEnv() error {
	output, err := getShellOutput("fetch", "docker compose ls --format json | jq '.[].Name' -r")
	if err != nil {
		return fmt.Errorf("failed to fetch running environments: %+v", err)
	}

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "mudly__") {
			name := strings.TrimPrefix(line, "mudly__")

			logrus.Infof("{%s} cleaning...", name)
			_, err := getShellOutput("fetch", fmt.Sprintf("docker compose -p mudly__%s down", name))
			if err != nil {
				logrus.Errorf("{%s} failed: %s", name, err.Error())
			} else {
				logrus.Infof("{%s}: done", name)
			}
		}
	}

	return nil
}
