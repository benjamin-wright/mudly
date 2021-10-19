package config

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"ponglehub.co.uk/tools/mudly/internal/target"
)

func getDirs(targets []target.Target) []string {
	dirs := []string{}

	for _, target := range targets {
		dupe := false
		for _, dir := range dirs {
			if target.Dir == dir {
				dupe = true
				break
			}
		}

		if !dupe {
			dirs = append(dirs, target.Dir)
		}
	}

	return dirs
}

func getDependencyTargets(config Config) []target.Target {
	newTargets := []target.Target{}

	for _, artefact := range config.Artefacts {
		for _, target := range artefact.DependsOn {
			newTargets = append(newTargets, config.Rebase(target))
		}

		parts := strings.Split(artefact.Pipeline, " ")
		if len(parts) == 2 {
			pipelineTarget := target.Target{Dir: parts[0], Artefact: "pipeline"}
			newTargets = append(newTargets, config.Rebase(pipelineTarget))
		}
	}

	return newTargets
}

func LoadConfigs(targets []target.Target) ([]Config, error) {
	configs := []Config{}
	for {
		newTargets := []target.Target{}

		for _, dir := range getDirs(targets) {
			got := false

			for _, cfg := range configs {
				if dir == cfg.Path {
					got = true
				}
			}

			if got {
				continue
			}

			cfg, err := getConfigData(dir)
			if err != nil {
				return nil, err
			}

			configs = append(configs, cfg)

			newTargets = append(newTargets, getDependencyTargets(cfg)...)
		}

		if len(newTargets) == 0 {
			break
		} else {
			targets = append(targets, newTargets...)
		}
	}

	return configs, nil
}

func getConfigData(filepath string) (Config, error) {
	r, isDir, err := openFile(filepath)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Path:  filepath,
		IsDir: isDir,
	}

	r.prune()

	for r.nextLine() {
		switch r.getLineType() {
		case ARTEFACT_LINE:
			artefact, err := getArtefact(r)
			if err != nil {
				return cfg, err
			}

			cfg.Artefacts = append(cfg.Artefacts, artefact)
		case PIPELINE_LINE:
			pipeline, err := getPipeline(r)
			if err != nil {
				return cfg, err
			}

			cfg.Pipelines = append(cfg.Pipelines, pipeline)
		case ENV_LINE:
			name, value, err := getEnv(r)
			if err != nil {
				return cfg, err
			}

			if cfg.Env == nil {
				cfg.Env = map[string]string{}
			}

			cfg.Env[name] = value
		case DOCKER_LINE:
			dockerfile, err := getDockerData(r)
			if err != nil {
				return cfg, err
			}

			if cfg.Dockerfile == nil {
				cfg.Dockerfile = []Dockerfile{}
			}

			cfg.Dockerfile = append(cfg.Dockerfile, dockerfile)
		case DEVENV_LINE:
			devenv, err := getDevEnvData(r)
			if err != nil {
				return cfg, err
			}

			if cfg.DevEnv == nil {
				cfg.DevEnv = []DevEnv{}
			}

			cfg.DevEnv = append(cfg.DevEnv, devenv)
		default:
			return cfg, fmt.Errorf("unknown line type: %s", r.line())
		}
	}

	return cfg, nil
}

func getDockerData(r *reader) (Dockerfile, error) {
	dockerfile := Dockerfile{}
	firstLine := r.line()

	trimmed := strings.TrimSpace(firstLine)
	parts := strings.Split(trimmed, " ")

	if len(parts) != 2 {
		return dockerfile, fmt.Errorf("failed to parse dockerfile line \"%s\", wrong number of arguments", firstLine)
	}

	dockerfile.Name = parts[1]

	targetIndent := r.indent()

	for r.nextLine() {
		indent := r.indent()

		if indent <= targetIndent {
			r.previousLine()
			break
		}

		switch r.getLineType() {
		case FILE_LINE:
			f, err := getStringOrMultiline(r, true)
			if err != nil {
				return dockerfile, fmt.Errorf("failed to parse dockerfile: %+v", err)
			}

			dockerfile.File = f
		case IGNORE_LINE:
			ignore, err := getStringOrMultiline(r, true)
			if err != nil {
				return dockerfile, fmt.Errorf("failed to parse dockerignore: %+v", err)
			}

			dockerfile.Ignore = ignore
		default:
			return dockerfile, fmt.Errorf("unknown line type: %s", r.line())
		}
	}

	return dockerfile, nil
}

func getDevEnvData(r *reader) (DevEnv, error) {
	devenv := DevEnv{}
	firstLine := r.line()

	trimmed := strings.TrimSpace(firstLine)
	parts := strings.Split(trimmed, " ")

	if len(parts) != 2 {
		return devenv, fmt.Errorf("failed to parse devenv line \"%s\", wrong number of arguments", firstLine)
	}

	devenv.Name = parts[1]

	targetIndent := r.indent()

	for r.nextLine() {
		indent := r.indent()

		if indent <= targetIndent {
			r.previousLine()
			break
		}

		switch r.getLineType() {
		case COMPOSE_LINE:
			f, err := getStringOrMultiline(r, true)
			if err != nil {
				return devenv, fmt.Errorf("failed to parse devenv: %+v", err)
			}

			devenv.Compose = f
		default:
			return devenv, fmt.Errorf("unknown line type: %s", r.line())
		}
	}

	return devenv, nil
}

func getArtefact(r *reader) (Artefact, error) {
	artefact := Artefact{}
	firstLine := r.line()

	trimmed := strings.TrimSpace(firstLine)
	parts := strings.Split(trimmed, " ")

	if len(parts) != 2 {
		return artefact, fmt.Errorf("failed to parse artefact line \"%s\", wrong number of arguments", firstLine)
	}

	artefact.Name = parts[1]

	targetIndent := r.indent()

	for r.nextLine() {
		indent := r.indent()

		if indent <= targetIndent {
			r.previousLine()
			break
		}

		switch r.getLineType() {
		case ENV_LINE:
			name, value, err := getEnv(r)
			if err != nil {
				return artefact, err
			}

			if artefact.Env == nil {
				artefact.Env = map[string]string{}
			}

			artefact.Env[name] = value
		case PIPELINE_LINE:
			name, err := getPipelineLink(r)
			if err != nil {
				return artefact, err
			}

			artefact.Pipeline = name
		case DEPENDS_LINE:
			t, err := getDepends(r)
			if err != nil {
				return artefact, err
			}

			if artefact.DependsOn == nil {
				artefact.DependsOn = []target.Target{}
			}

			artefact.DependsOn = append(artefact.DependsOn, t)
		case CONDITION_LINE:
			condition, err := getStringOrMultiline(r, false)
			if err != nil {
				return artefact, err
			}

			artefact.Condition = condition
		case DEVENV_LINE:
			devenv, err := getStringArg(r, nil)
			if err != nil {
				return artefact, err
			}

			artefact.DevEnv = devenv
		case STEP_LINE:
			step, err := getStep(r)
			if err != nil {
				return artefact, err
			}

			if artefact.Steps == nil {
				artefact.Steps = []Step{}
			}

			artefact.Steps = append(artefact.Steps, step)
		default:
			return artefact, fmt.Errorf("unknown line type: %s", r.line())
		}
	}

	return artefact, nil
}

func getPipeline(r *reader) (Pipeline, error) {
	pipeline := Pipeline{}
	firstLine := r.line()

	trimmed := strings.TrimSpace(firstLine)
	parts := strings.Split(trimmed, " ")

	if len(parts) != 2 {
		return pipeline, fmt.Errorf("failed to parse pipeline line \"%s\", wrong number of arguments", firstLine)
	}

	pipeline.Name = parts[1]

	targetIndent := r.indent()

	for r.nextLine() {
		indent := r.indent()

		if indent <= targetIndent {
			r.previousLine()
			break
		}

		switch r.getLineType() {
		case ENV_LINE:
			name, value, err := getEnv(r)
			if err != nil {
				return pipeline, err
			}

			if pipeline.Env == nil {
				pipeline.Env = map[string]string{}
			}

			pipeline.Env[name] = value
		case STEP_LINE:
			step, err := getStep(r)
			if err != nil {
				return pipeline, err
			}

			if pipeline.Steps == nil {
				pipeline.Steps = []Step{}
			}

			pipeline.Steps = append(pipeline.Steps, step)
		default:
			return pipeline, fmt.Errorf("unknown line type: %s", r.line())
		}
	}

	return pipeline, nil
}

func getPipelineLink(r *reader) (string, error) {
	trimmed := strings.TrimSpace(r.line())

	parts := strings.Split(trimmed, " ")

	if len(parts) < 2 || len(parts) > 3 {
		return "", fmt.Errorf("pipeline unknown syntax error for line \"%s\"", r.line())
	}

	return strings.Join(parts[1:], " "), nil
}

var envRegex *regexp.Regexp = regexp.MustCompile(`^(?:\s*)ENV (\S+)\=(\S+)$`)

func getEnv(r *reader) (string, string, error) {
	matches := envRegex.FindStringSubmatch(r.line())

	if matches == nil {
		return "", "", fmt.Errorf("env unknown syntax error for line \"%s\"", r.line())
	}

	if len(matches) != 3 {
		return "", "", fmt.Errorf("env match count error for line \"%s\" (found %d, expecting 2)", r.line(), len(matches)-1)
	}

	return matches[1], matches[2], nil
}

func getDepends(r *reader) (target.Target, error) {
	trimmed := strings.TrimSpace(r.line())

	parts := strings.Split(trimmed, " ")

	if len(parts) != 3 {
		return target.Target{}, fmt.Errorf("depends unknown syntax error for line \"%s\"", r.line())
	}

	t, err := target.ParseTarget(parts[2])
	if err != nil {
		return target.Target{}, err
	}

	if t == nil {
		return target.Target{}, fmt.Errorf("expected a target but got nil: \"%s\"", r.line())
	}

	return *t, nil
}

func getStep(r *reader) (Step, error) {
	step := Step{}
	firstLine := r.line()

	trimmed := strings.TrimSpace(firstLine)
	parts := strings.Split(trimmed, " ")

	if len(parts) != 2 {
		return step, fmt.Errorf("failed to parse artefact line \"%s\", wrong number of arguments", firstLine)
	}

	step.Name = parts[1]

	targetIndent := r.indent()

	for r.nextLine() {
		indent := r.indent()

		if indent <= targetIndent {
			r.previousLine()
			break
		}

		switch r.getLineType() {
		case ENV_LINE:
			name, value, err := getEnv(r)
			if err != nil {
				return step, err
			}

			if step.Env == nil {
				step.Env = map[string]string{}
			}

			step.Env[name] = value
		case WATCH_LINE:
			paths, err := getWatchPaths(r)
			if err != nil {
				return step, err
			}

			if step.Watch == nil {
				step.Watch = []string{}
			}

			step.Watch = append(step.Watch, paths...)
		case CONDITION_LINE:
			condition, err := getStringOrMultiline(r, false)
			if err != nil {
				return step, err
			}

			step.Condition = condition
		case COMMAND_LINE:
			command, err := getStringOrMultiline(r, false)
			if err != nil {
				return step, err
			}

			step.Command = command
		case DOCKER_LINE:
			dockerfile, err := getStringArg(r, nil)
			if err != nil {
				return step, err
			}

			step.Dockerfile = dockerfile
		case DEVENV_LINE:
			devenv, err := getStringArg(r, nil)
			if err != nil {
				return step, err
			}

			step.DevEnv = devenv
		case CONTEXT_LINE:
			context, err := getStringArg(r, nil)
			if err != nil {
				return step, err
			}

			step.Context = context
		case TAG_LINE:
			tag, err := getStringArg(r, nil)
			if err != nil {
				return step, err
			}

			step.Tag = tag
		case WAIT_FOR_LINE:
			waitFor, err := getStringArg(r, &getStringArgInputs{
				expectedLength:  3,
				acceptExtraArgs: true,
			})
			if err != nil {
				return step, err
			}

			if step.WaitFor == nil {
				step.WaitFor = []string{}
			}

			step.WaitFor = append(step.WaitFor, waitFor)
		default:
			return step, fmt.Errorf("unknown line type: %s", r.line())
		}
	}

	return step, nil
}

func getWatchPaths(r *reader) ([]string, error) {
	trimmed := strings.TrimSpace(r.line())

	parts := strings.Split(trimmed, " ")

	if len(parts) < 2 {
		return nil, fmt.Errorf("unknown syntax error for line \"%s\"", r.line())
	}

	return parts[1:], nil
}

type getStringArgInputs struct {
	expectedLength  int
	acceptExtraArgs bool
}

func getStringArg(r *reader, args *getStringArgInputs) (string, error) {
	if args == nil {
		args = &getStringArgInputs{
			expectedLength: 2,
		}
	}

	trimmed := strings.TrimSpace(r.line())
	parts := strings.Split(trimmed, " ")

	if args.acceptExtraArgs {
		if len(parts) < args.expectedLength {
			return "", fmt.Errorf("unknown syntax error for line \"%s\"", r.line())
		}

		return strings.Join(parts[args.expectedLength-1:], " "), nil
	} else {
		if len(parts) != args.expectedLength {
			return "", fmt.Errorf("unknown syntax error for line \"%s\"", r.line())
		}

		return parts[args.expectedLength-1], nil
	}
}

func getStringOrMultiline(r *reader, ignoreFirstLine bool) (string, error) {
	trimmed := strings.TrimSpace(r.line())

	parts := strings.Split(trimmed, " ")

	if len(parts) > 1 && !ignoreFirstLine {
		return strings.Join(parts[1:], " "), nil
	}

	lines := []string{}
	targetIndent := r.indent()
	steppedIndent := targetIndent

	for r.nextLine() {
		indent := r.indent()

		if indent <= targetIndent {
			r.previousLine()
			break
		}

		if steppedIndent == targetIndent {
			steppedIndent = indent
		}

		lines = append(lines, r.line()[steppedIndent:])
	}

	if len(lines) == 0 {
		return "", errors.New("empty string / multiline-string not supported")
	}

	return strings.Join(lines, "\n"), nil
}
