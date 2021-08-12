package steps

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"ponglehub.co.uk/tools/mudly/internal/runner"
	"ponglehub.co.uk/tools/mudly/internal/utils"
)

type CommandStep struct {
	Name      string
	Condition string
	Watch     []string
	Command   string
	Env       map[string]string
	WaitFor   []string
}

type Checker interface {
	SaveTimestamp(dir string, artefact string, step string) error
	FetchTimestamp(dir string, artefact string, step string) (time.Time, error)
	HasChangedSince(t time.Time, paths []string) (bool, error)
}

func SetMockChecker(instance Checker) {
	ageCheckerInstance = instance
}

var ageCheckerInstance Checker = &utils.AgeChecker{}

func (c CommandStep) Run(dir string, artefact string, env map[string]string) runner.CommandResult {
	merged := utils.MergeMaps(env, c.Env)

	if c.Watch != nil && len(c.Watch) > 0 {
		t, err := ageCheckerInstance.FetchTimestamp(dir, artefact, c.Name)
		if err != nil {
			logrus.Errorf("%s[%s]: failed fetching timestamp: %+v", artefact, c.Name, err)
			return runner.COMMAND_ERROR
		}

		hasChanged, err := ageCheckerInstance.HasChangedSince(t, c.Watch)
		if err != nil {
			logrus.Errorf("%s[%s]: failed comparing timestamp: %+v", artefact, c.Name, err)
			return runner.COMMAND_ERROR
		}

		if !hasChanged {
			logrus.Debugf("%s[%s (test)]: Skipping step with no changes", artefact, c.Name)
			return runner.COMMAND_SKIPPED
		}
	}

	if c.Condition != "" {
		test := runShellCommand(&shellCommand{
			dir:      dir,
			artefact: artefact,
			step:     fmt.Sprintf("%s (test)", c.Name),
			command:  "/bin/bash",
			args:     []string{"-c", c.Condition},
			env:      merged,
			test:     true,
		})

		if !test {
			logrus.Debugf("%s[%s (test)]: Skipping step, condition not met", artefact, c.Name)
			return runner.COMMAND_SKIPPED
		}
	}

	for idx, waitFor := range c.WaitFor {
		for {
			test := runShellCommand(&shellCommand{
				dir:      dir,
				artefact: artefact,
				step:     fmt.Sprintf("%s (wait:%d)", c.Name, idx),
				command:  "/bin/bash",
				args:     []string{"-c", waitFor},
				env:      merged,
				test:     true,
			})

			if test {
				break
			} else {
				time.Sleep(time.Millisecond * 500)
				logrus.Debugf("%s[%s (wait:%d)]: Wait for failed, trying again...", artefact, c.Name, idx)
			}
		}
	}

	success := runShellCommand(&shellCommand{
		dir:      dir,
		artefact: artefact,
		step:     c.Name,
		command:  "/bin/bash",
		args:     []string{"-c", c.Command},
		env:      merged,
	})

	if success {
		logrus.Debugf("writing timestamp...")
		if err := ageCheckerInstance.SaveTimestamp(dir, artefact, c.Name); err != nil {
			logrus.Warnf("Failed to write timestamp: %+v", err)
		}
		return runner.COMMAND_SUCCESS
	} else {
		return runner.COMMAND_ERROR
	}
}

func (c CommandStep) String() string {
	return c.Name
}
