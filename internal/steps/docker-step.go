package steps

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"ponglehub.co.uk/tools/mudly/internal/runner"
)

type DockerStep struct {
	Name         string
	Dockerfile   string
	Dockerignore string
	Context      string
	Tag          string
}

func (d DockerStep) args() []string {
	args := []string{"build"}

	if d.Tag != "" {
		args = append(args, "-t", d.Tag)
	}

	args = append(args, "-f", "-")

	if d.Context != "" {
		args = append(args, d.Context)
	} else {
		args = append(args, ".")
	}

	return args
}

func (d DockerStep) Run(dir string, artefact string, env map[string]string) runner.CommandResult {
	if d.Dockerignore != "" {
		if err := ioutil.WriteFile(path.Join(dir, ".dockerignore"), []byte(d.Dockerignore), 0644); err != nil {
			logrus.Errorf("%s[%s]: Failed to write .dockerignore: %+v", artefact, d.Name, err)
			return runner.COMMAND_ERROR
		}

		defer func() {
			if err := os.Remove(path.Join(dir, ".dockerignore")); err != nil {
				logrus.Errorf("%s[%s]: Failed to clean up .dockerignore: %+v", artefact, d.Name, err)
			}
		}()
	}

	success := runShellCommand(&shellCommand{
		dir:      dir,
		artefact: artefact,
		step:     d.Name,
		command:  "docker",
		args:     d.args(),
		stdin:    d.Dockerfile,
	})

	if success {
		return runner.COMMAND_SUCCESS
	} else {
		return runner.COMMAND_ERROR
	}
}

func (d DockerStep) String() string {
	return d.Name
}
