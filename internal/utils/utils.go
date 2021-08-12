package utils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func MergeMaps(maps ...map[string]string) map[string]string {
	output_map := map[string]string{}

	hasAny := false
	for _, obj := range maps {
		for key, value := range obj {
			output_map[key] = value
			hasAny = true
		}
	}

	if hasAny {
		return output_map
	} else {
		return nil
	}
}

type AgeChecker struct{}

func (a *AgeChecker) HasChangedSince(t time.Time, paths []string) (bool, error) {
	workdir, err := osInstance.Getwd()
	if err != nil {
		return false, fmt.Errorf("failed to get working dir: %+v", err)
	}

	for _, p := range paths {
		matches, err := filepath.Glob(filepath.Join(workdir, p))
		if err != nil {
			return false, err
		}

		for _, match := range matches {
			stat, err := os.Stat(match)
			if err != nil {
				return false, err
			}

			if stat.IsDir() {
				continue
			}

			if stat.ModTime().After(t) {
				return true, nil
			}
		}
	}

	return false, nil
}

func (a *AgeChecker) FetchTimestamp(config string, artefact string, step string) (time.Time, error) {
	mudlyDir, ok := os.LookupEnv("MUDLY_DIR")
	if !ok {
		mudlyDir = os.ExpandEnv("$HOME/.mudly")
	}

	data, err := os.ReadFile(path.Join(mudlyDir, "timestamps"))
	if err != nil {
		return time.Time{}, err
	}

	timestampData := TimestampData{}
	err = yaml.Unmarshal(data, &timestampData)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp data: %+v", err)
	}

	workdir, err := osInstance.Getwd()
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get working dir: %+v", err)
	}

	for _, c := range timestampData.Configs {
		if c.Path == path.Join(workdir, config) {
			for _, a := range c.Artefacts {
				if a.Name == artefact {
					for _, s := range a.Steps {
						if s.Name == step {
							return time.Unix(s.Timestamp, 0), nil
						}
					}
				}
			}
		}
	}

	return time.Now(), nil
}

func (a *AgeChecker) SaveTimestamp(config string, artefact string, step string) error {
	mudlyDir, ok := os.LookupEnv("MUDLY_DIR")
	if !ok {
		mudlyDir = os.ExpandEnv("$HOME/.mudly")
	}

	err := os.MkdirAll(mudlyDir, 0766)
	if err != nil {
		return fmt.Errorf("failed to make directory chain: %+v", err)
	}

	data, err := os.ReadFile(path.Join(mudlyDir, "timestamps"))
	if err != nil {
		data = []byte{}
	}

	timestampData := TimestampData{}
	err = yaml.Unmarshal(data, &timestampData)
	if err != nil {
		timestampData.Configs = []Config{}
	}

	updatedData, err := updateConfig(timestampData, config, artefact, step)
	if err != nil {
		return err
	}

	outData, err := yaml.Marshal(&updatedData)
	if err != nil {
		return err
	}

	os.WriteFile(path.Join(mudlyDir, "timestamps"), outData, 0766)

	return nil
}

type OSTools interface {
	Getwd() (string, error)
	GetTimestamp() int64
}

type osImpl struct{}

func (o *osImpl) Getwd() (string, error) { return os.Getwd() }
func (o *osImpl) GetTimestamp() int64    { return time.Now().Unix() }

var osInstance OSTools = &osImpl{}

func updateConfig(timestampData TimestampData, config string, artefact string, step string) (TimestampData, error) {
	workdir, err := osInstance.Getwd()
	if err != nil {
		return timestampData, fmt.Errorf("failed to get working dir: %+v", err)
	}

	cfgAbsPath := path.Join(workdir, config)
	cfgIndex := -1

	for index, c := range timestampData.Configs {
		if c.Path == cfgAbsPath {
			cfgIndex = index
		}
	}

	if cfgIndex == -1 {
		timestampData.Configs = append(timestampData.Configs, Config{
			Path: cfgAbsPath,
		})

		cfgIndex = len(timestampData.Configs) - 1
	}

	artefactIndex := -1
	for index, a := range timestampData.Configs[cfgIndex].Artefacts {
		if a.Name == artefact {
			artefactIndex = index
		}
	}

	if artefactIndex == -1 {
		timestampData.Configs[cfgIndex].Artefacts = append(timestampData.Configs[cfgIndex].Artefacts, Artefact{
			Name: artefact,
		})

		artefactIndex = len(timestampData.Configs[cfgIndex].Artefacts) - 1
	}

	stepIndex := -1
	for index, s := range timestampData.Configs[cfgIndex].Artefacts[artefactIndex].Steps {
		if s.Name == step {
			stepIndex = index
		}
	}

	if stepIndex == -1 {
		timestampData.Configs[cfgIndex].Artefacts[artefactIndex].Steps = append(timestampData.Configs[cfgIndex].Artefacts[artefactIndex].Steps, Step{
			Name: step,
		})

		stepIndex = len(timestampData.Configs[cfgIndex].Artefacts[artefactIndex].Steps) - 1
	}

	logrus.Debugf("Indices: %d, %d, %d", cfgIndex, artefactIndex, stepIndex)

	timestampData.Configs[cfgIndex].Artefacts[artefactIndex].Steps[stepIndex].Timestamp = osInstance.GetTimestamp()

	return timestampData, nil
}

type Step struct {
	Name      string `yaml:"name"`
	Timestamp int64  `yaml:"timestamp"`
}

type Artefact struct {
	Name  string `yaml:"name"`
	Steps []Step `yaml:"steps"`
}

type Config struct {
	Path      string     `yaml:"path"`
	Artefacts []Artefact `yaml:"artefacts"`
}

type TimestampData struct {
	Configs []Config `yaml:"configs"`
}
