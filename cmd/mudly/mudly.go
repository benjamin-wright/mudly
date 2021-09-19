package main

import (
	"flag"
	"os"

	"github.com/sirupsen/logrus"
	"ponglehub.co.uk/tools/mudly/internal/args_parser"
	"ponglehub.co.uk/tools/mudly/internal/config"
	"ponglehub.co.uk/tools/mudly/internal/runner"
	"ponglehub.co.uk/tools/mudly/internal/solver"
	"ponglehub.co.uk/tools/mudly/internal/steps"
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

func buildTargets(options args_parser.Options) {
	logrus.Debugf("Targets: %+v", options.Targets)

	configs, err := config.LoadConfigs(options.Targets)
	if err != nil {
		logrus.Fatalf("Error loading config: %+v", err)
	}

	logrus.Debugf("Configs: %+v", configs)

	var stripTargets []target.Target
	if options.OnlyDeps {
		stripTargets = options.Targets
	}

	nodes, err := solver.Solve(&solver.SolveInputs{
		Targets:      options.Targets,
		Configs:      configs,
		StripTargets: stripTargets,
		NoDeps:       options.NoDeps,
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

func stop() {
	err := steps.CleanupDevEnv()
	if err != nil {
		logrus.Fatalf(err.Error())
	}
}

func main() {
	setLogLevel()

	args := os.Args[1:]
	logrus.Debugf("Running mudly with args: %+v", args)

	command, options, err := args_parser.Parse(args)

	switch command {
	case args_parser.BUILD_COMMAND:
		buildTargets(options)
	case args_parser.STOP_COMMAND:
		stop()
	case args_parser.NO_COMMAND:
		logrus.Fatalf(err.Error())
	}
}
