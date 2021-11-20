package steps

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"ponglehub.co.uk/tools/mudly/internal/runner"
	"ponglehub.co.uk/tools/mudly/internal/utils"
)

type DockerStep struct {
	Name         string
	Dockerfile   string
	Dockerignore string
	Context      string
	Tag          string
	BuildArg     map[string]string
}

func (d DockerStep) args(env map[string]string) []string {
	args := []string{"build"}

	if d.Tag != "" {
		args = append(args, "-t", d.Tag)
	}

	args = append(args, "-f", "-")

	merged := utils.MergeMaps(env, d.BuildArg)

	for key, value := range merged {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}

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
			logrus.Errorf("{%s} %s[%s]: Failed to write .dockerignore: %+v", dir, artefact, d.Name, err)
			return runner.COMMAND_ERROR
		}

		defer func() {
			if err := os.Remove(path.Join(dir, ".dockerignore")); err != nil {
				logrus.Errorf("{%s} %s[%s]: Failed to clean up .dockerignore: %+v", dir, artefact, d.Name, err)
			}
		}()
	}

	success := runShellCommand(&shellCommand{
		dir:      dir,
		artefact: artefact,
		step:     d.Name,
		command:  "docker",
		args:     d.args(env),
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
