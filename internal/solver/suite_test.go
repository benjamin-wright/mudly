package solver_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"ponglehub.co.uk/tools/mudly/internal/config"
	"ponglehub.co.uk/tools/mudly/internal/runner"
	"ponglehub.co.uk/tools/mudly/internal/solver"
	"ponglehub.co.uk/tools/mudly/internal/steps"
	"ponglehub.co.uk/tools/mudly/internal/target"
)

type testNode struct {
	Path      string
	Artefact  string
	SharedEnv map[string]string
	Step      runner.Runnable
	State     runner.NodeState
	DependsOn []int
}

func convert(nodes []testNode) []*runner.Node {
	converted := []*runner.Node{}

	for id := range nodes {
		node := nodes[id]
		converted = append(converted, &runner.Node{
			Path:      node.Path,
			Artefact:  node.Artefact,
			SharedEnv: node.SharedEnv,
			Step:      node.Step,
			State:     node.State,
			DependsOn: []*runner.Node{},
		})
	}

	for idx, node := range nodes {
		for _, dep := range node.DependsOn {
			converted[idx].DependsOn = append(converted[idx].DependsOn, converted[dep])
		}
	}

	return converted
}

func TestSolver(t *testing.T) {
	for _, test := range []struct {
		Name         string
		Targets      []target.Target
		StripTargets []target.Target
		Configs      []config.Config
		NoDeps       bool
		Expected     []testNode
	}{
		{
			Name:    "simple",
			Targets: []target.Target{{Dir: ".", Artefact: "image"}},
			Configs: []config.Config{
				{
					Path: ".",
					Dockerfile: []config.Dockerfile{
						{Name: "my-image", File: "dockerfile contents"},
					},
					Artefacts: []config.Artefact{
						{
							Name: "image",
							Steps: []config.Step{
								{
									Name:    "build",
									Command: "go build -o ./bin/mudly ./cmd/mudly",
								},
								{
									Name:       "docker",
									Dockerfile: "my-image",
									Context:    ".",
								},
							},
						},
					},
				},
			},
			Expected: []testNode{
				{
					Path:     ".",
					Artefact: "image",
					Step: steps.CommandStep{
						Name:    "build",
						Command: "go build -o ./bin/mudly ./cmd/mudly",
					},
					State:     runner.STATE_PENDING,
					DependsOn: []int{},
				},
				{
					Path:     ".",
					Artefact: "image",
					Step: steps.DockerStep{
						Name:       "docker",
						Dockerfile: "dockerfile contents",
						Context:    ".",
					},
					State:     runner.STATE_PENDING,
					DependsOn: []int{0},
				},
			},
		},
		{
			Name:    "parallel builds",
			Targets: []target.Target{{Dir: ".", Artefact: "image"}, {Dir: ".", Artefact: "something"}},
			Configs: []config.Config{
				{
					Path: ".",
					Dockerfile: []config.Dockerfile{
						{Name: "my-image", File: "my dockerfile contents"},
					},
					Artefacts: []config.Artefact{
						{
							Name: "image",
							Steps: []config.Step{
								{
									Name:    "build",
									Command: "go build -o ./bin/mudly ./cmd/mudly",
								},
								{
									Name:       "docker",
									Dockerfile: "my-image",
									Context:    ".",
								},
							},
						},
						{
							Name: "something",
							Steps: []config.Step{
								{
									Name:    "echo",
									Command: "echo \"hi\"",
								},
								{
									Name:    "build",
									Command: "whatevs",
								},
							},
						},
					},
				},
			},
			Expected: []testNode{
				{
					Path:     ".",
					Artefact: "image",
					Step: steps.CommandStep{
						Name:    "build",
						Command: "go build -o ./bin/mudly ./cmd/mudly",
					},
					State:     runner.STATE_PENDING,
					DependsOn: []int{},
				},
				{
					Path:     ".",
					Artefact: "image",
					Step: steps.DockerStep{
						Name:       "docker",
						Dockerfile: "my dockerfile contents",
						Context:    ".",
					},
					State:     runner.STATE_PENDING,
					DependsOn: []int{0},
				},
				{
					Path:     ".",
					Artefact: "something",
					Step: steps.CommandStep{
						Name:    "echo",
						Command: "echo \"hi\"",
					},
					State:     runner.STATE_PENDING,
					DependsOn: []int{},
				},
				{
					Path:     ".",
					Artefact: "something",
					Step: steps.CommandStep{
						Name:    "build",
						Command: "whatevs",
					},
					State:     runner.STATE_PENDING,
					DependsOn: []int{2},
				},
			},
		},
		{
			Name:    "linked builds",
			Targets: []target.Target{{Dir: ".", Artefact: "image"}},
			Configs: []config.Config{
				{
					Path: ".",
					Env: map[string]string{
						"GLOBAL_ENV": "value3",
					},
					Dockerfile: []config.Dockerfile{
						{Name: "image-1", File: "image 1 content"},
					},
					Pipelines: []config.Pipeline{
						{
							Name: "my-pipeline",
							Env: map[string]string{
								"PIPELINE_ENV": "value1",
							},
							Steps: []config.Step{
								{
									Name:    "build",
									Command: "go build -o ./bin/mudly ./cmd/mudly",
									Env: map[string]string{
										"STEP_ENV": "value0",
									},
								},
								{
									Name:       "docker",
									Dockerfile: "image-1",
									Context:    ".",
								},
							},
						},
					},
					Artefacts: []config.Artefact{
						{
							Name: "image",
							Env: map[string]string{
								"ARTEFACT_ENV": "value2",
							},
							DependsOn: []target.Target{
								{Dir: ".", Artefact: "something"},
							},
							Pipeline: "my-pipeline",
						},
						{
							Name: "something",
							Steps: []config.Step{
								{
									Name:    "echo",
									Command: "echo \"hi\"",
								},
								{
									Name:    "build",
									Command: "whatevs",
								},
							},
						},
					},
				},
			},
			Expected: []testNode{
				{
					Path:     ".",
					Artefact: "image",
					SharedEnv: map[string]string{
						"PIPELINE_ENV": "value1",
						"ARTEFACT_ENV": "value2",
						"GLOBAL_ENV":   "value3",
					},
					Step: steps.CommandStep{
						Name:    "build",
						Command: "go build -o ./bin/mudly ./cmd/mudly",
						Env: map[string]string{
							"STEP_ENV": "value0",
						},
					},
					State:     runner.STATE_PENDING,
					DependsOn: []int{3},
				},
				{
					Path:     ".",
					Artefact: "image",
					SharedEnv: map[string]string{
						"PIPELINE_ENV": "value1",
						"ARTEFACT_ENV": "value2",
						"GLOBAL_ENV":   "value3",
					},
					Step: steps.DockerStep{
						Name:       "docker",
						Dockerfile: "image 1 content",
						Context:    ".",
					},
					State:     runner.STATE_PENDING,
					DependsOn: []int{0},
				},
				{
					Path:     ".",
					Artefact: "something",
					SharedEnv: map[string]string{
						"GLOBAL_ENV": "value3",
					},
					Step: steps.CommandStep{
						Name:    "echo",
						Command: "echo \"hi\"",
					},
					State:     runner.STATE_PENDING,
					DependsOn: []int{},
				},
				{
					Path:     ".",
					Artefact: "something",
					SharedEnv: map[string]string{
						"GLOBAL_ENV": "value3",
					},
					Step: steps.CommandStep{
						Name:    "build",
						Command: "whatevs",
					},
					State:     runner.STATE_PENDING,
					DependsOn: []int{2},
				},
			},
		},
		{
			Name:         "dependency builds",
			Targets:      []target.Target{{Dir: ".", Artefact: "image"}},
			StripTargets: []target.Target{{Dir: ".", Artefact: "image"}},
			Configs: []config.Config{
				{
					Path: ".",
					Env: map[string]string{
						"GLOBAL_ENV": "value3",
					},
					Dockerfile: []config.Dockerfile{
						{Name: "image-1", File: "image 1 content"},
					},
					Pipelines: []config.Pipeline{
						{
							Name: "my-pipeline",
							Env: map[string]string{
								"PIPELINE_ENV": "value1",
							},
							Steps: []config.Step{
								{
									Name:    "build",
									Command: "go build -o ./bin/mudly ./cmd/mudly",
									Env: map[string]string{
										"STEP_ENV": "value0",
									},
								},
								{
									Name:       "docker",
									Dockerfile: "image-1",
									Context:    ".",
								},
							},
						},
					},
					Artefacts: []config.Artefact{
						{
							Name: "image",
							Env: map[string]string{
								"ARTEFACT_ENV": "value2",
							},
							DependsOn: []target.Target{
								{Dir: ".", Artefact: "something"},
							},
							Pipeline: "my-pipeline",
						},
						{
							Name: "something",
							Steps: []config.Step{
								{
									Name:    "echo",
									Command: "echo \"hi\"",
								},
								{
									Name:    "build",
									Command: "whatevs",
								},
							},
						},
					},
				},
			},
			Expected: []testNode{
				{
					Path:     ".",
					Artefact: "something",
					SharedEnv: map[string]string{
						"GLOBAL_ENV": "value3",
					},
					Step: steps.CommandStep{
						Name:    "echo",
						Command: "echo \"hi\"",
					},
					State:     runner.STATE_PENDING,
					DependsOn: []int{},
				},
				{
					Path:     ".",
					Artefact: "something",
					SharedEnv: map[string]string{
						"GLOBAL_ENV": "value3",
					},
					Step: steps.CommandStep{
						Name:    "build",
						Command: "whatevs",
					},
					State:     runner.STATE_PENDING,
					DependsOn: []int{0},
				},
			},
		},
		{
			Name:         "stepless artefact",
			Targets:      []target.Target{{Dir: ".", Artefact: "image"}},
			StripTargets: []target.Target{{Dir: ".", Artefact: "image"}},
			Configs: []config.Config{
				{
					Path: ".",
					Env: map[string]string{
						"GLOBAL_ENV": "value3",
					},
					Dockerfile: []config.Dockerfile{
						{Name: "image-1", File: "image 1 content"},
					},
					Artefacts: []config.Artefact{
						{
							Name: "image",
							Env: map[string]string{
								"ARTEFACT_ENV": "value2",
							},
							DependsOn: []target.Target{
								{Dir: ".", Artefact: "something"},
							},
						},
						{
							Name: "something",
							Steps: []config.Step{
								{
									Name:    "echo",
									Command: "echo \"hi\"",
								},
								{
									Name:    "build",
									Command: "whatevs",
								},
							},
						},
					},
				},
			},
			Expected: []testNode{
				{
					Path:     ".",
					Artefact: "something",
					SharedEnv: map[string]string{
						"GLOBAL_ENV": "value3",
					},
					Step: steps.CommandStep{
						Name:    "echo",
						Command: "echo \"hi\"",
					},
					State:     runner.STATE_PENDING,
					DependsOn: []int{},
				},
				{
					Path:     ".",
					Artefact: "something",
					SharedEnv: map[string]string{
						"GLOBAL_ENV": "value3",
					},
					Step: steps.CommandStep{
						Name:    "build",
						Command: "whatevs",
					},
					State:     runner.STATE_PENDING,
					DependsOn: []int{0},
				},
			},
		},
		{
			Name:    "remote-pipeline",
			Targets: []target.Target{{Dir: "subdir", Artefact: "image"}},
			Configs: []config.Config{
				{
					Path: "subdir",
					Artefacts: []config.Artefact{
						{
							Name:     "image",
							Pipeline: "../otherdir remote-pipeline",
						},
					},
				},
				{
					Path: "otherdir",
					Dockerfile: []config.Dockerfile{
						{Name: "my-image", File: "dockerfile contents"},
					},
					Pipelines: []config.Pipeline{
						{
							Name: "remote-pipeline",
							Steps: []config.Step{
								{
									Name:    "build",
									Command: "go build -o ./bin/mudly ./cmd/mudly",
								},
								{
									Name:       "docker",
									Dockerfile: "my-image",
									Context:    ".",
								},
							},
						},
					},
				},
			},
			Expected: []testNode{
				{
					Path:     "subdir",
					Artefact: "image",
					Step: steps.CommandStep{
						Name:    "build",
						Command: "go build -o ./bin/mudly ./cmd/mudly",
					},
					State:     runner.STATE_PENDING,
					DependsOn: []int{},
				},
				{
					Path:     "subdir",
					Artefact: "image",
					Step: steps.DockerStep{
						Name:       "docker",
						Dockerfile: "dockerfile contents",
						Context:    ".",
					},
					State:     runner.STATE_PENDING,
					DependsOn: []int{0},
				},
			},
		},
		{
			Name:    "no deps",
			Targets: []target.Target{{Dir: ".", Artefact: "image"}},
			NoDeps:  true,
			Configs: []config.Config{
				{
					Path: ".",
					Env: map[string]string{
						"GLOBAL_ENV": "value3",
					},
					Dockerfile: []config.Dockerfile{
						{Name: "image-1", File: "image 1 content"},
					},
					Pipelines: []config.Pipeline{
						{
							Name: "my-pipeline",
							Env: map[string]string{
								"PIPELINE_ENV": "value1",
							},
							Steps: []config.Step{
								{
									Name:    "build",
									Command: "go build -o ./bin/mudly ./cmd/mudly",
									Env: map[string]string{
										"STEP_ENV": "value0",
									},
								},
								{
									Name:       "docker",
									Dockerfile: "image-1",
									Context:    ".",
								},
							},
						},
					},
					Artefacts: []config.Artefact{
						{
							Name: "image",
							Env: map[string]string{
								"ARTEFACT_ENV": "value2",
							},
							DependsOn: []target.Target{
								{Dir: ".", Artefact: "something"},
							},
							Pipeline: "my-pipeline",
						},
						{
							Name: "something",
							Steps: []config.Step{
								{
									Name:    "echo",
									Command: "echo \"hi\"",
								},
								{
									Name:    "build",
									Command: "whatevs",
								},
							},
						},
					},
				},
			},
			Expected: []testNode{
				{
					Path:     ".",
					Artefact: "image",
					SharedEnv: map[string]string{
						"PIPELINE_ENV": "value1",
						"ARTEFACT_ENV": "value2",
						"GLOBAL_ENV":   "value3",
					},
					Step: steps.CommandStep{
						Name:    "build",
						Command: "go build -o ./bin/mudly ./cmd/mudly",
						Env: map[string]string{
							"STEP_ENV": "value0",
						},
					},
					State: runner.STATE_PENDING,
				},
				{
					Path:     ".",
					Artefact: "image",
					SharedEnv: map[string]string{
						"PIPELINE_ENV": "value1",
						"ARTEFACT_ENV": "value2",
						"GLOBAL_ENV":   "value3",
					},
					Step: steps.DockerStep{
						Name:       "docker",
						Dockerfile: "image 1 content",
						Context:    ".",
					},
					State:     runner.STATE_PENDING,
					DependsOn: []int{0},
				},
			},
		},
		{
			Name:    "devenv",
			Targets: []target.Target{{Dir: ".", Artefact: "image"}},
			Configs: []config.Config{
				{
					Path: ".",
					DevEnv: []config.DevEnv{
						{
							Name:    "myenv",
							Compose: "composefile",
						},
					},
					Artefacts: []config.Artefact{
						{
							Name:   "image",
							DevEnv: "myenv",
							Steps: []config.Step{
								{
									Name:    "build",
									Command: "go build -o ./bin/mudly ./cmd/mudly",
								},
								{
									Name:    "docker",
									Command: "thing 2",
								},
							},
						},
					},
				},
			},
			Expected: []testNode{
				{
					Path:     ".",
					Artefact: "image",
					Step: steps.DevenvStep{
						Name:    "myenv",
						Compose: "composefile",
					},
					State:     runner.STATE_PENDING,
					DependsOn: []int{},
				},
				{
					Path:     ".",
					Artefact: "image",
					Step: steps.CommandStep{
						Name:    "build",
						Command: "go build -o ./bin/mudly ./cmd/mudly",
					},
					State:     runner.STATE_PENDING,
					DependsOn: []int{0},
				},
				{
					Path:     ".",
					Artefact: "image",
					Step: steps.CommandStep{
						Name:    "docker",
						Command: "thing 2",
					},
					State:     runner.STATE_PENDING,
					DependsOn: []int{1},
				},
			},
		},
		{
			Name:    "env priority",
			Targets: []target.Target{{Dir: ".", Artefact: "image"}},
			NoDeps:  true,
			Configs: []config.Config{
				{
					Path: ".",
					Pipelines: []config.Pipeline{
						{
							Name: "my-pipeline",
							Env: map[string]string{
								"SAME_VAR": "pipeline-val",
							},
							Steps: []config.Step{
								{
									Name:    "build",
									Command: "go build -o ./bin/mudly ./cmd/mudly",
								},
							},
						},
					},
					Artefacts: []config.Artefact{
						{
							Name: "image",
							Env: map[string]string{
								"SAME_VAR": "artefact-val",
							},
							Pipeline: "my-pipeline",
						},
					},
				},
			},
			Expected: []testNode{
				{
					Path:     ".",
					Artefact: "image",
					SharedEnv: map[string]string{
						"SAME_VAR": "artefact-val",
					},
					Step: steps.CommandStep{
						Name:    "build",
						Command: "go build -o ./bin/mudly ./cmd/mudly",
					},
					State: runner.STATE_PENDING,
				},
			},
		},
	} {
		t.Run(test.Name, func(u *testing.T) {
			nodes, err := solver.Solve(&solver.SolveInputs{
				Targets:      test.Targets,
				Configs:      test.Configs,
				StripTargets: test.StripTargets,
				NoDeps:       test.NoDeps,
			})

			assert.NoError(u, err, "didn't expect an error")

			if test.Expected != nil {
				if nodes != nil {
					assert.Equal(u, convert(test.Expected), nodes)
				} else {
					assert.Fail(u, "expected a list of nodes")
				}
			}
		})
	}
}
