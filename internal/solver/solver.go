package solver

import (
	"fmt"
	"path"
	"strings"

	"ponglehub.co.uk/tools/mudly/internal/config"
	"ponglehub.co.uk/tools/mudly/internal/runner"
	"ponglehub.co.uk/tools/mudly/internal/steps"
	"ponglehub.co.uk/tools/mudly/internal/target"
	"ponglehub.co.uk/tools/mudly/internal/utils"
)

type link struct {
	Target target.Target
	Source target.Target
}

func (l link) isSame(m link) bool {
	return l.Source.IsSame(m.Source) && l.Target.IsSame(m.Target)
}

func getArtefact(target target.Target, configs []config.Config) (*config.Config, *config.Artefact, error) {
	var cfg config.Config
	missing := true
	for _, c := range configs {
		if path.Clean(target.Dir) == c.Path {
			cfg = c
			missing = false
			break
		}
	}

	if missing {
		return nil, nil, fmt.Errorf("couldn't find target config %s", target.Dir)
	}

	var artefact config.Artefact
	missing = true
	for _, a := range cfg.Artefacts {
		if a.Name == target.Artefact {
			artefact = a
			missing = false
			break
		}
	}

	if missing {
		return nil, nil, fmt.Errorf("couldn't find target artefact %s+%s", target.Dir, target.Artefact)
	}

	return &cfg, &artefact, nil
}

func getPipeline(configs []config.Config, cfg *config.Config, artefact *config.Artefact) (*config.Config, *config.Pipeline, error) {
	if artefact.Steps != nil && len(artefact.Steps) > 0 {
		return cfg, &config.Pipeline{
			Name:  "",
			Steps: artefact.Steps,
		}, nil
	} else if artefact.Pipeline != "" {
		if strings.Contains(artefact.Pipeline, " ") {
			parts := strings.Split(artefact.Pipeline, " ")
			pipelineTarget := target.Target{Dir: parts[0], Artefact: parts[1]}
			pipelineTarget = cfg.Rebase(pipelineTarget)

			for _, c := range configs {
				if c.Path == pipelineTarget.Dir {
					for _, pipeline := range c.Pipelines {
						if pipeline.Name == pipelineTarget.Artefact {
							return &c, &pipeline, nil
						}
					}
				}
			}
		} else {
			for _, pipeline := range cfg.Pipelines {
				if pipeline.Name == artefact.Pipeline {
					return cfg, &pipeline, nil
				}
			}
		}

		return nil, nil, fmt.Errorf("failed to get pipeline from artefact %s (%s)", artefact.Name, cfg.Path)
	}

	return nil, nil, nil
}

func collectDependencies(targets []target.Target, configs []config.Config) ([]link, error) {
	running := true
	links := []link{}

	for running {
		newLinks := []link{}
		newTargets := []target.Target{}

		for _, target := range targets {
			cfg, artefact, err := getArtefact(target, configs)
			if err != nil {
				return nil, err
			}

			for _, dep := range artefact.DependsOn {
				rebased := cfg.Rebase(dep)

				link := link{
					Target: rebased,
					Source: target,
				}

				missing := true

				for _, existing := range links {
					if link.isSame(existing) {
						missing = false
						break
					}
				}

				if missing {
					newLinks = append(newLinks, link)
					newTargets = append(newTargets, rebased)
				}
			}
		}

		if len(newLinks) == 0 {
			running = false
			continue
		}

		links = append(links, newLinks...)
		targets = append(targets, newTargets...)
	}

	output := []link{}

	for _, link := range links {
		missing := true

		for _, existing := range output {
			if link.isSame(existing) {
				missing = false
				break
			}
		}

		if missing {
			output = append(output, link)
		}
	}

	return output, nil
}

func getDedupedTargets(targets []target.Target, links []link) []target.Target {
	for _, link := range links {
		targets = append(targets, link.Target)
	}

	output := []target.Target{}

	for _, target := range targets {
		missing := true

		for _, existing := range output {
			if target.IsSame(existing) {
				missing = false
			}
		}

		if missing {
			output = append(output, target)
		}
	}

	return output
}

func pruneLinks(links []link, configs []config.Config) []link {
	keepers := append([]link{}, links...)
	running := true

	for running {
		running = false
		inputs := append([]link{}, keepers...)
		keepers = []link{}

		for _, l := range inputs {
			var sourceArtefact *config.Artefact

			for _, c := range configs {
				if c.Path == l.Source.Dir {
					for id, a := range c.Artefacts {
						if a.Name == l.Source.Artefact {
							sourceArtefact = &c.Artefacts[id]
						}
					}
				}
			}

			if sourceArtefact == nil || (len(sourceArtefact.Steps) == 0 && sourceArtefact.Pipeline == "") {
				continue
			}

			var targetArtefact *config.Artefact

			for _, c := range configs {
				if c.Path == l.Target.Dir {
					for id, a := range c.Artefacts {
						if a.Name == l.Target.Artefact {
							targetArtefact = &c.Artefacts[id]
						}
					}
				}
			}

			if targetArtefact == nil {
				continue
			}

			if len(targetArtefact.Steps) > 0 || targetArtefact.Pipeline != "" {
				keepers = append(keepers, l)
			} else {
				running = true
				source := l.Source
				for _, t := range targetArtefact.DependsOn {
					keepers = append(keepers, link{
						Source: target.Target{Dir: source.Dir, Artefact: source.Artefact},
						Target: target.Target{Dir: t.Dir, Artefact: t.Artefact},
					})
				}
			}
		}
	}

	return keepers
}

func getStrippedTargets(deduped []target.Target, stripTargets []target.Target) []target.Target {
	stripped := []target.Target{}
	for _, d := range deduped {
		keep := true
		for _, t := range stripTargets {
			if d.Dir == t.Dir && d.Artefact == t.Artefact {
				keep = false
			}
		}

		if keep {
			stripped = append(stripped, d)
		}
	}

	return stripped
}

func createRunnable(step config.Step) (runner.Runnable, error) {
	if step.Command != "" {
		return steps.CommandStep{
			Name:      step.Name,
			Condition: step.Condition,
			Watch:     step.Watch,
			Command:   step.Command,
			Env:       step.Env,
			WaitFor:   step.WaitFor,
		}, nil
	}

	if step.Dockerfile != "" {
		return steps.DockerStep{
			Name:         step.Name,
			Dockerfile:   step.Dockerfile,
			Dockerignore: step.Ignore,
			Context:      step.Context,
			Tag:          step.Tag,
		}, nil
	}

	return nil, fmt.Errorf("failed to convert config step into runnable step: %+v", step)
}

func createNodes(targets []target.Target, configs []config.Config) (*NodeList, error) {
	nodes := NodeList{list: []nodeListElement{}}

	for _, target := range targets {
		cfg, artefact, err := getArtefact(target, configs)
		if err != nil {
			return &nodes, err
		}

		if artefact.Condition != "" {
			nodes.AddNode(cfg.Path, artefact.Name, &runner.Node{
				SharedEnv: utils.MergeMaps(cfg.Env, artefact.Env),
				Path:      cfg.Path,
				Artefact:  artefact.Name,
				Step:      steps.ArtefactStep{Condition: artefact.Condition},
				State:     runner.STATE_PENDING,
				DependsOn: []*runner.Node{},
			})
		}

		pipelineConfig, pipeline, err := getPipeline(configs, cfg, artefact)
		if err != nil {
			return &nodes, err
		}
		if pipeline == nil {
			continue
		}

		if artefact.DevEnv != "" {
			var d *config.DevEnv
			for index, devenv := range pipelineConfig.DevEnv {
				if devenv.Name == artefact.DevEnv {
					d = &pipelineConfig.DevEnv[index]
					break
				}
			}

			if d == nil {
				return &nodes, fmt.Errorf("failed to get devenv for reference: %s", artefact.DevEnv)
			}

			newNode := runner.Node{
				SharedEnv: utils.MergeMaps(cfg.Env, pipeline.Env, artefact.Env),
				Path:      cfg.Path,
				Artefact:  artefact.Name,
				Step: steps.DevenvStep{
					Name:    d.Name,
					Compose: d.Compose,
				},
				State:     runner.STATE_PENDING,
				DependsOn: []*runner.Node{},
			}

			nodes.AddNode(cfg.Path, artefact.Name, &newNode)
		}

		for _, step := range pipeline.Steps {
			if step.DevEnv != "" {
				var d *config.DevEnv
				for index, devenv := range pipelineConfig.DevEnv {
					if devenv.Name == step.DevEnv {
						d = &pipelineConfig.DevEnv[index]
						break
					}
				}

				if d == nil {
					return &nodes, fmt.Errorf("failed to get devenv for reference: %s", step.DevEnv)
				}

				newNode := runner.Node{
					SharedEnv: utils.MergeMaps(cfg.Env, pipeline.Env, artefact.Env),
					Path:      cfg.Path,
					Artefact:  artefact.Name,
					Step: steps.DevenvStep{
						Name:    d.Name,
						Compose: d.Compose,
					},
					State:     runner.STATE_PENDING,
					DependsOn: []*runner.Node{},
				}

				nodes.AddNode(cfg.Path, artefact.Name, &newNode)
			}

			if step.Dockerfile != "" {
				content := ""
				ignore := ""
				for _, dockerfile := range pipelineConfig.Dockerfile {
					if dockerfile.Name == step.Dockerfile {
						content = dockerfile.File
						ignore = dockerfile.Ignore
					}
				}

				if content != "" {
					step.Dockerfile = content
					step.Ignore = ignore
				} else {
					return &nodes, fmt.Errorf("failed to get dockerfile for reference: %s", step.Dockerfile)
				}
			}

			runnable, err := createRunnable(step)
			if err != nil {
				return &nodes, err
			}

			newNode := runner.Node{
				SharedEnv: utils.MergeMaps(cfg.Env, pipeline.Env, artefact.Env),
				Path:      cfg.Path,
				Artefact:  artefact.Name,
				Step:      runnable,
				State:     runner.STATE_PENDING,
				DependsOn: []*runner.Node{},
			}

			nodes.AddNode(cfg.Path, artefact.Name, &newNode)
		}
	}

	return &nodes, nil
}

func isStripTarget(dir string, artefact string, stripTargets []target.Target) bool {
	for _, t := range stripTargets {
		if t.Dir == dir && t.Artefact == artefact {
			return true
		}
	}

	return false
}

func linkNodes(links []link, nodes *NodeList, stripTargets []target.Target) error {
	for _, link := range links {
		sourceNode := nodes.getFirstElement(link.Source.Dir, link.Source.Artefact)
		if sourceNode == nil {
			if isStripTarget(link.Source.Dir, link.Source.Artefact, stripTargets) {
				continue
			}

			return fmt.Errorf("failed to generate link for %+v, couldn't find source element", link)
		}

		targetNode := nodes.getLastElement(link.Target.Dir, link.Target.Artefact)
		if targetNode == nil {
			if isStripTarget(link.Source.Dir, link.Source.Artefact, stripTargets) {
				continue
			}

			return fmt.Errorf("failed to generate link for %+v, couldn't find target element", link)
		}

		sourceNode.node.DependsOn = append(sourceNode.node.DependsOn, targetNode.node)
	}

	return nil
}

type SolveInputs struct {
	Targets      []target.Target
	Configs      []config.Config
	StripTargets []target.Target
	NoDeps       bool
}

func Solve(inputs *SolveInputs) ([]*runner.Node, error) {
	// Recursively compile the chain of dependency links between the input targets and their references
	// and their references references.
	links, err := collectDependencies(inputs.Targets, inputs.Configs)
	if err != nil {
		return nil, err
	}

	// Reduce the target and dependency list down to just unique config and artefact combinations
	deduped := getDedupedTargets(inputs.Targets, links)

	// Remove any targets that were only meant for dependency gathering
	stripped := getStrippedTargets(deduped, inputs.StripTargets)

	// Remove links to upstream artefacts that don't have any steps
	pruned := pruneLinks(links, inputs.Configs)

	if inputs.NoDeps {
		stripped = inputs.Targets
		pruned = nil
	}

	// Create the solver node list for all the unique config and artefact combinations
	nodes, err := createNodes(stripped, inputs.Configs)
	if err != nil {
		return nil, err
	}

	// Decorate the node list with the dependency links, so that we can figure out the build order
	err = linkNodes(pruned, nodes, inputs.StripTargets)
	if err != nil {
		return nil, err
	}

	return nodes.GetList(), nil
}
