package target

import (
	"errors"
	"fmt"
	"path"
	"strings"
)

type Target struct {
	Dir      string
	Artefact string
}

func (t Target) Rebase(target Target) Target {
	return Target{
		Dir:      path.Clean(fmt.Sprintf("%s/%s", target.Dir, t.Dir)),
		Artefact: t.Artefact,
	}
}

func (t Target) IsSame(u Target) bool {
	return t.Dir == u.Dir && t.Artefact == u.Artefact
}

func ParseTarget(filepath string) (*Target, error) {
	if filepath == "" {
		return nil, errors.New("failed to parse target with empty string")
	}

	parts := strings.Split(filepath, "+")
	if len(parts) != 2 {
		return nil, fmt.Errorf("failed to parse target from path: %s", filepath)
	}

	dir := parts[0]
	artefact := parts[1]

	if dir == "" {
		dir = "."
	}

	return &Target{
		Dir:      path.Clean(dir),
		Artefact: artefact,
	}, nil
}
