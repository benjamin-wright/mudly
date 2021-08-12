package config

import (
	"errors"
	"fmt"
	"strings"
)

type reader struct {
	lines []string
	index int
}

func openFile(filepath string) (*reader, error) {
	data, err := fsInstance.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %+v", err)
	}

	if strings.Contains(string(data), "\t") {
		return nil, errors.New("your file has tabs in it")
	}

	lines := strings.Split(string(data), "\n")
	index := -1

	return &reader{
		lines: lines,
		index: index,
	}, nil
}

func (r *reader) prune() {
	output := []string{}

	for _, line := range r.lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		output = append(output, line)
	}

	r.lines = output
}

type lineType int

const (
	ARTEFACT_LINE lineType = iota
	PIPELINE_LINE
	ENV_LINE
	DEPENDS_LINE
	STEP_LINE
	WATCH_LINE
	CONDITION_LINE
	COMMAND_LINE
	DOCKER_LINE
	FILE_LINE
	IGNORE_LINE
	CONTEXT_LINE
	TAG_LINE
	WAIT_FOR_LINE
	UNKNOWN_LINE
	READER_ERROR
)

func (r *reader) getLineType() lineType {
	if r.index < 0 || r.index >= len(r.lines) {
		return READER_ERROR
	}

	trimmed := strings.TrimSpace(r.lines[r.index])

	if strings.HasPrefix(trimmed, "ENV") {
		return ENV_LINE
	}

	if strings.HasPrefix(trimmed, "ARTEFACT") {
		return ARTEFACT_LINE
	}

	if strings.HasPrefix(trimmed, "DEPENDS ON") {
		return DEPENDS_LINE
	}

	if strings.HasPrefix(trimmed, "STEP") {
		return STEP_LINE
	}

	if strings.HasPrefix(trimmed, "WATCH") {
		return WATCH_LINE
	}

	if strings.HasPrefix(trimmed, "CONDITION") {
		return CONDITION_LINE
	}

	if strings.HasPrefix(trimmed, "COMMAND") {
		return COMMAND_LINE
	}

	if strings.HasPrefix(trimmed, "DOCKERFILE") {
		return DOCKER_LINE
	}

	if strings.HasPrefix(trimmed, "FILE") {
		return FILE_LINE
	}

	if strings.HasPrefix(trimmed, "IGNORE") {
		return IGNORE_LINE
	}

	if strings.HasPrefix(trimmed, "CONTEXT") {
		return CONTEXT_LINE
	}

	if strings.HasPrefix(trimmed, "TAG") {
		return TAG_LINE
	}

	if strings.HasPrefix(trimmed, "WAIT FOR") {
		return WAIT_FOR_LINE
	}

	if strings.HasPrefix(trimmed, "PIPELINE") {
		return PIPELINE_LINE
	}

	return UNKNOWN_LINE
}

func (r *reader) indent() int {
	if r.index >= len(r.lines) {
		return -1
	}

	line := r.line()
	trimmed := strings.TrimLeft(line, " ")

	return len(line) - len(trimmed)
}

func (r *reader) nextLine() bool {
	if r.index >= len(r.lines)-1 {
		return false
	}

	r.index++
	return true
}

func (r *reader) previousLine() {
	if r.index > 0 {
		r.index--
	}
}

func (r *reader) line() string {
	if r.index < 0 || r.index >= len(r.lines) {
		return ""
	}

	return r.lines[r.index]
}
