package config

import (
	"fmt"
	"path"

	"ponglehub.co.uk/tools/mudly/internal/target"
)

type Dockerfile struct {
	Name   string
	File   string
	Ignore string
}

type DevEnv struct {
	Name    string
	Compose string
}

type Config struct {
	Path       string
	IsDir      bool
	DevEnv     []DevEnv
	Dockerfile []Dockerfile
	Artefacts  []Artefact
	Pipelines  []Pipeline
	Env        map[string]string
}

func (c *Config) WorkDir() string {
	if !c.IsDir {
		return path.Dir(c.Path)
	}

	return c.Path
}

func (c *Config) Rebase(t target.Target) target.Target {
	if c.IsDir || t.Dir == "." {
		return target.Target{
			Dir:      path.Clean(fmt.Sprintf("%s/%s", c.Path, t.Dir)),
			Artefact: t.Artefact,
		}
	}

	return target.Target{
		Dir:      path.Clean(fmt.Sprintf("%s/%s", path.Dir(c.Path), t.Dir)),
		Artefact: t.Artefact,
	}
}

// func (c *Config) IsMatch(filepath string) bool {
// 	targetDir := path.Clean(filepath)

// 	if c.Name == "Mudfile" {
// 		return targetDir == c.Path
// 	} else {
// 		return path.Clean(fmt.Sprintf("%s.Mudfile", targetDir)) == path.Clean(fmt.Sprintf("%s/%s", c.Path, c.Name))
// 	}
// }

// func (c *Config) Retarget(t target.Target) target.Target {
// 	if c.Name != "Mudfile" {
// 		if t.Dir == "." {
// 			t.Dir = strings.Replace(c.Name, ".Mudfile", "", 1)
// 		}
// 	}

// 	return t
// }

type Pipeline struct {
	Name  string
	Env   map[string]string
	Steps []Step
}

type Artefact struct {
	Name      string
	DependsOn []target.Target
	Env       map[string]string
	Steps     []Step
	Pipeline  string
	Condition string
	DevEnv    string
}

type Step struct {
	Name       string
	Env        map[string]string
	Condition  string
	Command    string
	Context    string
	DevEnv     string
	Watch      []string
	Dockerfile string
	Ignore     string
	Tag        string
	WaitFor    []string
}
