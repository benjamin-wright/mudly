package steps

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

type shellCommand struct {
	dir      string
	artefact string
	step     string
	command  string
	args     []string
	env      map[string]string
	test     bool
	stdin    string
}

func runShellCommand(command *shellCommand) bool {
	logrus.Debugf("{%s} %s[%s]: Running command %s %s", command.dir, command.artefact, command.step, command.command, strings.Join(command.args, " "))

	cmd := exec.Command(command.command, command.args...)
	cmd.Dir = path.Clean(command.dir)

	cmd.Env = os.Environ()
	for key, value := range command.env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logrus.Errorf("{%s} %s[%s]: Failed to get command stdout pipe: %+v", command.dir, command.artefact, command.step, err)
		return false
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		logrus.Errorf("{%s} %s[%s]: Failed to get command stderr pipe: %+v", command.dir, command.artefact, command.step, err)
		return false
	}

	if command.stdin != "" {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			logrus.Errorf("{%s} %s[%s]: Failed to get command stdin pipe: %+v", command.dir, command.artefact, command.step, err)
		}

		go func() {
			defer stdin.Close()
			io.WriteString(stdin, command.stdin)
		}()
	}

	err = cmd.Start()
	if err != nil {
		logrus.Errorf("{%s} %s[%s]: Command couldn't run: %+v", command.dir, command.artefact, command.step, err)
		return false
	}

	go func() {
		stdoutScanner := bufio.NewScanner(stdout)
		stdoutScanner.Split(bufio.ScanLines)
		for stdoutScanner.Scan() {
			logrus.Infof("{%s} %s[%s]: %s", command.dir, command.artefact, command.step, stdoutScanner.Text())
		}
	}()

	go func() {
		stderrScanner := bufio.NewScanner(stderr)
		stderrScanner.Split(bufio.ScanLines)
		for stderrScanner.Scan() {
			logrus.Warnf("{%s} %s[%s]: %s", command.dir, command.artefact, command.step, stderrScanner.Text())
		}
	}()

	err = cmd.Wait()
	if err != nil {
		if !command.test {
			logrus.Errorf("{%s} %s[%s]: Command failed: %+v", command.dir, command.artefact, command.step, err)
		}
		return false
	}

	return true
}
