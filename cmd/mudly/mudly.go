package main

import (
	"flag"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"ponglehub.co.uk/tools/mudly/internal/config"
	"ponglehub.co.uk/tools/mudly/internal/runner"
	"ponglehub.co.uk/tools/mudly/internal/solver"
	"ponglehub.co.uk/tools/mudly/internal/target"
)

func setLogLevel() {
	if logLevel, ok := os.LookupEnv("MUDLY_LOG_LEVEL"); ok {
		parsedLevel, err := logrus.ParseLevel(logLevel)
		if err != nil {
			logrus.Fatalf("Failed to parse log level %s from environment", logLevel)
		}

		logrus.SetLevel(parsedLevel)
	} else if logLevel := flag.String("log-level", "info", "the logging level to use"); logLevel != nil {
		parsedLevel, err := logrus.ParseLevel(*logLevel)
		if err != nil {
			logrus.Fatalf("Failed to parse log level %s from --log-level flag", *logLevel)
		}

		logrus.SetLevel(parsedLevel)
	}
}

type CommandType int

const (
	NONE_COMMAND CommandType = iota
	DEPS_COMMAND
	NO_DEPS_COMMAND
)

func getCommand(args []string) (CommandType, []string) {
	if len(args) == 0 {
		logrus.Fatalf("must provide a build target or command")
	}

	if strings.Contains(args[0], "+") {
		return NONE_COMMAND, args
	}

	switch args[0] {
	case "deps", "dependencies":
		return DEPS_COMMAND, args[1:]
	case "no-deps", "no-dependencies":
		return NO_DEPS_COMMAND, args[1:]
	default:
		logrus.Fatalf("must provide a valid build target or command")
		panic("logrus fatal should exit")
	}
}

func buildTargets(command CommandType, args []string) {
	targets := []target.Target{}
	for _, path := range args {
		target, err := target.ParseTarget(path)
		if err != nil {
			logrus.Fatalf("Error parsing target: %+v", err)
		}

		targets = append(targets, *target)
	}

	logrus.Debugf("Targets: %+v", targets)

	configs, err := config.LoadConfigs(targets)
	if err != nil {
		logrus.Fatalf("Error loading config: %+v", err)
	}

	logrus.Debugf("Configs: %+v", configs)

	var stripTargets []target.Target
	if command == DEPS_COMMAND {
		stripTargets = targets
	}

	nodes, err := solver.Solve(&solver.SolveInputs{
		Targets:      targets,
		Configs:      configs,
		StripTargets: stripTargets,
		NoDeps:       command == NO_DEPS_COMMAND,
	})
	if err != nil {
		logrus.Fatalf("Error in solver: %+v", err)
	}

	if len(nodes) == 0 {
		logrus.Info("Nothing to build")
		return
	}

	for _, node := range nodes {
		logrus.Debugf("Node: %+v", *node)
	}

	err = runner.Run(nodes)
	if err != nil {
		logrus.Fatalf("Error in runner: %+v", err)
	}

	for _, node := range nodes {
		logrus.Debugf("%s:%s[%s] - %d", node.Path, node.Artefact, node.Step, node.State)
	}
}

func main() {
	setLogLevel()

	args := os.Args[1:]
	logrus.Debugf("Running mudly with args: %+v", args)

	command, args := getCommand(args)

	buildTargets(command, args)
}
