package solver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"ponglehub.co.uk/tools/mudly/internal/config"
	"ponglehub.co.uk/tools/mudly/internal/target"
)

type getArtefactResult struct {
	Config   string
	Artefact string
	Pipeline string
}

func TestGetArtefact(t *testing.T) {
	for _, test := range []struct {
		Name     string
		Target   target.Target
		Configs  []config.Config
		Expected *getArtefactResult
	}{
		{
			Name:   "simple",
			Target: target.Target{Dir: ".", Artefact: "test-artefact"},
			Configs: []config.Config{
				{
					Path: ".",
					Artefacts: []config.Artefact{
						{
							Name: "test-artefact",
						},
					},
				},
			},
			Expected: &getArtefactResult{Config: ".", Artefact: "test-artefact"},
		},
		{
			Name:   "picks the right artefact",
			Target: target.Target{Dir: ".", Artefact: "other"},
			Configs: []config.Config{
				{
					Path: ".",
					Artefacts: []config.Artefact{
						{
							Name: "test-artefact",
						},
						{
							Name: "other",
						},
					},
				},
			},
			Expected: &getArtefactResult{Config: ".", Artefact: "other"},
		},
		{
			Name:   "picks the right artefact reverse",
			Target: target.Target{Dir: ".", Artefact: "other"},
			Configs: []config.Config{
				{
					Path: ".",
					Artefacts: []config.Artefact{
						{
							Name: "other",
						},
						{
							Name: "test-artefact",
						},
					},
				},
			},
			Expected: &getArtefactResult{Config: ".", Artefact: "other"},
		},
		{
			Name:   "picks the right config",
			Target: target.Target{Dir: "./subdir", Artefact: "test-artefact"},
			Configs: []config.Config{
				{
					Path: ".",
					Artefacts: []config.Artefact{
						{
							Name:     "test-artefact",
							Pipeline: "firstConfig",
						},
					},
				},
				{
					Path: "subdir",
					Artefacts: []config.Artefact{
						{
							Name:     "test-artefact",
							Pipeline: "secondConfig",
						},
					},
				},
			},
			Expected: &getArtefactResult{Config: "subdir", Artefact: "test-artefact", Pipeline: "secondConfig"},
		},
		{
			Name:   "picks the right config reversed",
			Target: target.Target{Dir: "./subdir", Artefact: "test-artefact"},
			Configs: []config.Config{
				{
					Path: "subdir",
					Artefacts: []config.Artefact{
						{
							Name:     "test-artefact",
							Pipeline: "firstConfig",
						},
					},
				},
				{
					Path: ".",
					Artefacts: []config.Artefact{
						{
							Name:     "test-artefact",
							Pipeline: "secondConfig",
						},
					},
				},
			},
			Expected: &getArtefactResult{Config: "subdir", Artefact: "test-artefact", Pipeline: "firstConfig"},
		},
	} {
		t.Run(test.Name, func(u *testing.T) {
			cfg, artefact, err := getArtefact(test.Target, test.Configs)

			assert.NoError(u, err, "didn't expect an error")

			if test.Expected != nil {
				if cfg != nil && artefact != nil {
					assert.Equal(u, test.Expected, &getArtefactResult{Config: cfg.Path, Artefact: artefact.Name, Pipeline: artefact.Pipeline})
				} else {
					assert.Fail(u, "expected a config and artefact", "%+v, %+v", cfg, artefact)
				}
			}
		})
	}
}
func TestGetPipeline(t *testing.T) {
	for _, test := range []struct {
		Name     string
		Configs  []config.Config
		Config   *config.Config
		Artefact *config.Artefact
		Expected *config.Pipeline
		Error    string
	}{
		{
			Name: "take artefact steps",
			Artefact: &config.Artefact{
				Steps: []config.Step{
					{Name: "hi"},
					{Name: "ho"},
				},
			},
			Expected: &config.Pipeline{Steps: []config.Step{{Name: "hi"}, {Name: "ho"}}},
		},
		{
			Name: "take pipeline steps if artefact has none",
			Config: &config.Config{
				Pipelines: []config.Pipeline{
					{Name: "wrong-one", Steps: []config.Step{{Name: "hi"}}},
					{Name: "pipeline-name", Steps: []config.Step{{Name: "ho"}}},
				},
			},
			Artefact: &config.Artefact{
				Pipeline: "pipeline-name",
			},
			Expected: &config.Pipeline{Name: "pipeline-name", Steps: []config.Step{{Name: "ho"}}},
		},
		{
			Name: "take pipeline steps from remote reference",
			Configs: []config.Config{
				{
					Path:  "subdir",
					IsDir: true,
					Pipelines: []config.Pipeline{
						{Name: "pipeline-name", Steps: []config.Step{{Name: "not me"}}},
					},
				},
				{
					Path:  "otherdir",
					IsDir: true,
					Pipelines: []config.Pipeline{
						{Name: "wrong-one", Steps: []config.Step{{Name: "hi"}}},
						{Name: "pipeline-name", Steps: []config.Step{{Name: "ho"}}},
					},
				},
			},
			Config: &config.Config{
				Path:  "subdir",
				IsDir: true,
				Pipelines: []config.Pipeline{
					{Name: "pipeline-name", Steps: []config.Step{{Name: "not me"}}},
				},
			},
			Artefact: &config.Artefact{
				Pipeline: "../otherdir pipeline-name",
			},
			Expected: &config.Pipeline{Name: "pipeline-name", Steps: []config.Step{{Name: "ho"}}},
		},
		{
			Name: "error if pipeline not found",
			Config: &config.Config{
				Path:  "some-dir",
				IsDir: true,
				Pipelines: []config.Pipeline{
					{Name: "pipeline-name", Steps: []config.Step{{Name: "ho"}}},
				},
			},
			Artefact: &config.Artefact{
				Name:     "my-artefact",
				Pipeline: "wrong-name",
			},
			Error: "failed to get pipeline from artefact my-artefact (some-dir)",
		},
	} {
		t.Run(test.Name, func(u *testing.T) {
			_, pipeline, err := getPipeline(test.Configs, test.Config, test.Artefact)

			if test.Error != "" {
				assert.EqualError(u, err, test.Error)
			} else {
				assert.NoError(u, err, "didn't expect an error")
			}

			if test.Expected != nil {
				if pipeline != nil {
					assert.Equal(u, test.Expected, pipeline)
				} else {
					assert.Fail(u, "expected a pipeline", "%+v", pipeline)
				}
			}
		})
	}
}

func TestCollectDependencies(t *testing.T) {
	for _, test := range []struct {
		Name     string
		Targets  []target.Target
		Configs  []config.Config
		Expected []link
	}{
		{
			Name: "should get nothing from nothing",
		},
		{
			Name: "should find local links",
			Targets: []target.Target{
				{Dir: ".", Artefact: "artefact-1"},
			},
			Configs: []config.Config{
				{
					Path:  ".",
					IsDir: true,
					Artefacts: []config.Artefact{
						{
							Name: "artefact-1",
							DependsOn: []target.Target{
								{Dir: ".", Artefact: "artefact-2"},
							},
						},
						{
							Name: "artefact-2",
						},
					},
				},
			},
			Expected: []link{
				{
					Target: target.Target{Dir: ".", Artefact: "artefact-2"},
					Source: target.Target{Dir: ".", Artefact: "artefact-1"},
				},
			},
		},
		{
			Name: "should find remote links",
			Targets: []target.Target{
				{Dir: "subdir1", Artefact: "artefact-1"},
			},
			Configs: []config.Config{
				{
					Path:  "subdir1",
					IsDir: true,
					Artefacts: []config.Artefact{
						{
							Name: "artefact-1",
							DependsOn: []target.Target{
								{Dir: "../subdir2", Artefact: "artefact-2"},
							},
						},
						{
							Name: "artefact-2",
						},
					},
				},
				{
					Path:  "subdir2",
					IsDir: true,
					Artefacts: []config.Artefact{
						{
							Name: "artefact-2",
						},
					},
				},
			},
			Expected: []link{
				{
					Target: target.Target{Dir: "subdir2", Artefact: "artefact-2"},
					Source: target.Target{Dir: "subdir1", Artefact: "artefact-1"},
				},
			},
		},
		{
			Name: "should find chained dependency links",
			Targets: []target.Target{
				{Dir: "subdir1", Artefact: "artefact-1"},
			},
			Configs: []config.Config{
				{
					Path:  "subdir1",
					IsDir: true,
					Artefacts: []config.Artefact{
						{
							Name: "artefact-1",
							DependsOn: []target.Target{
								{Dir: "../subdir2", Artefact: "artefact-2"},
							},
						},
						{
							Name: "artefact-2",
						},
					},
				},
				{
					Path:  "subdir2",
					IsDir: true,
					Artefacts: []config.Artefact{
						{
							Name: "artefact-2",
							DependsOn: []target.Target{
								{Dir: "../subdir1", Artefact: "artefact-2"},
							},
						},
					},
				},
			},
			Expected: []link{
				{
					Target: target.Target{Dir: "subdir2", Artefact: "artefact-2"},
					Source: target.Target{Dir: "subdir1", Artefact: "artefact-1"},
				},
				{
					Target: target.Target{Dir: "subdir1", Artefact: "artefact-2"},
					Source: target.Target{Dir: "subdir2", Artefact: "artefact-2"},
				},
			},
		},
	} {
		t.Run(test.Name, func(u *testing.T) {
			links, err := collectDependencies(test.Targets, test.Configs)

			assert.NoError(u, err, "didn't expect an error")

			if test.Expected != nil {
				if links != nil {
					assert.Equal(u, test.Expected, links)
				} else {
					assert.Fail(u, "expected a list of links")
				}
			}
		})
	}
}

func TestGetDedupedTargets(t *testing.T) {
	for _, test := range []struct {
		Name     string
		Targets  []target.Target
		Links    []link
		Expected []target.Target
	}{
		{
			Name: "should get nothing from nothing",
		},
		{
			Name: "should return non-duplicated targets",
			Targets: []target.Target{
				{Dir: ".", Artefact: "artefact-1"},
				{Dir: ".", Artefact: "artefact-2"},
			},
			Expected: []target.Target{
				{Dir: ".", Artefact: "artefact-1"},
				{Dir: ".", Artefact: "artefact-2"},
			},
		},
		{
			Name: "should add linked targets",
			Targets: []target.Target{
				{Dir: ".", Artefact: "artefact-1"},
				{Dir: ".", Artefact: "artefact-2"},
			},
			Links: []link{
				{
					Source: target.Target{Dir: ".", Artefact: "artefact-1"},
					Target: target.Target{Dir: ".", Artefact: "artefact-3"},
				},
				{
					Source: target.Target{Dir: ".", Artefact: "artefact-2"},
					Target: target.Target{Dir: "subdir", Artefact: "artefact-1"},
				},
			},
			Expected: []target.Target{
				{Dir: ".", Artefact: "artefact-1"},
				{Dir: ".", Artefact: "artefact-2"},
				{Dir: ".", Artefact: "artefact-3"},
				{Dir: "subdir", Artefact: "artefact-1"},
			},
		},
		{
			Name: "should eliminate input and linked duplicates",
			Targets: []target.Target{
				{Dir: ".", Artefact: "artefact-1"},
				{Dir: ".", Artefact: "artefact-2"},
				{Dir: ".", Artefact: "artefact-2"},
				{Dir: ".", Artefact: "artefact-3"},
			},
			Links: []link{
				{
					Source: target.Target{Dir: ".", Artefact: "artefact-1"},
					Target: target.Target{Dir: ".", Artefact: "artefact-2"},
				},
				{
					Source: target.Target{Dir: ".", Artefact: "artefact-1"},
					Target: target.Target{Dir: "subdir", Artefact: "artefact-1"},
				},
				{
					Source: target.Target{Dir: ".", Artefact: "artefact-3"},
					Target: target.Target{Dir: "subdir", Artefact: "artefact-1"},
				},
			},
			Expected: []target.Target{
				{Dir: ".", Artefact: "artefact-1"},
				{Dir: ".", Artefact: "artefact-2"},
				{Dir: ".", Artefact: "artefact-3"},
				{Dir: "subdir", Artefact: "artefact-1"},
			},
		},
	} {
		t.Run(test.Name, func(u *testing.T) {
			deduped := getDedupedTargets(test.Targets, test.Links)

			if test.Expected != nil {
				if deduped != nil {
					assert.Equal(u, test.Expected, deduped)
				} else {
					assert.Fail(u, "expected a list of targets")
				}
			}
		})
	}
}

func TestPruneLinks(t *testing.T) {
	// To allow step-less artefacts properly, update the prune-links phase to not just eliminate links that reference stepless source artefacts,
	// but to bridge gaps, i.e.
	// artefact1 -> artefact2(stepless) -> artefact3 becomes artefact1 -> artefact3

	for _, test := range []struct {
		Name     string
		Links    []link
		Configs  []config.Config
		Expected []link
	}{
		{
			Name:     "empty in, empty out",
			Links:    []link{},
			Configs:  []config.Config{},
			Expected: []link{},
		},
		{
			Name: "same config",
			Links: []link{
				{
					Source: target.Target{Dir: "subdir", Artefact: "a1"},
					Target: target.Target{Dir: "subdir", Artefact: "a2"},
				},
			},
			Configs: []config.Config{
				{Path: "subdir", Artefacts: []config.Artefact{{Name: "a1", Pipeline: "pipeline"}, {Name: "a2", Pipeline: "pipeline"}}},
			},
			Expected: []link{
				{
					Source: target.Target{Dir: "subdir", Artefact: "a1"},
					Target: target.Target{Dir: "subdir", Artefact: "a2"},
				},
			},
		},
		{
			Name: "different configs",
			Links: []link{
				{
					Source: target.Target{Dir: "subdir1", Artefact: "a1"},
					Target: target.Target{Dir: "subdir2", Artefact: "a2"},
				},
			},
			Configs: []config.Config{
				{Path: "subdir1", Artefacts: []config.Artefact{{Name: "a1", Pipeline: "pipeline"}}},
				{Path: "subdir2", Artefacts: []config.Artefact{{Name: "a2", Pipeline: "pipeline"}}},
			},
			Expected: []link{
				{
					Source: target.Target{Dir: "subdir1", Artefact: "a1"},
					Target: target.Target{Dir: "subdir2", Artefact: "a2"},
				},
			},
		},
		{
			Name: "prune stepless source",
			Links: []link{
				{
					Source: target.Target{Dir: "subdir1", Artefact: "a1"},
					Target: target.Target{Dir: "subdir2", Artefact: "a2"},
				},
			},
			Configs: []config.Config{
				{Path: "subdir1", Artefacts: []config.Artefact{{Name: "a1"}}},
				{Path: "subdir2", Artefacts: []config.Artefact{{Name: "a2", Pipeline: "pipeline"}}},
			},
			Expected: []link{},
		},
		{
			Name: "prune stepless target",
			Links: []link{
				{
					Source: target.Target{Dir: "subdir1", Artefact: "a1"},
					Target: target.Target{Dir: "subdir2", Artefact: "a2"},
				},
			},
			Configs: []config.Config{
				{Path: "subdir1", Artefacts: []config.Artefact{{Name: "a1", Pipeline: "pipeline"}}},
				{Path: "subdir2", Artefacts: []config.Artefact{{Name: "a2"}}},
			},
			Expected: []link{},
		},
		{
			Name: "bridge stepless to dependencies",
			Links: []link{
				{
					Source: target.Target{Dir: "subdir1", Artefact: "c1"},
					Target: target.Target{Dir: "subdir2", Artefact: "c2"},
				},
			},
			Configs: []config.Config{
				{Path: "subdir1", Artefacts: []config.Artefact{{Name: "c1", Pipeline: "pipeline"}}},
				{
					Path: "subdir2",
					Artefacts: []config.Artefact{
						{
							Name: "c2",
							DependsOn: []target.Target{
								{Dir: "subdir2", Artefact: "d1"},
								{Dir: "subdir2", Artefact: "d2"},
							},
						},
						{
							Name:     "d1",
							Pipeline: "pipeline",
							DependsOn: []target.Target{
								{Dir: "subdir2", Artefact: "d3"},
							},
						},
						{
							Name: "d2",
							DependsOn: []target.Target{
								{Dir: "subdir2", Artefact: "d4"},
							},
						},
						{Name: "d3", Pipeline: "pipeline"},
						{Name: "d4", Pipeline: "pipeline"},
					},
				},
			},
			Expected: []link{
				{
					Source: target.Target{Dir: "subdir1", Artefact: "c1"},
					Target: target.Target{Dir: "subdir2", Artefact: "d1"},
				},
				{
					Source: target.Target{Dir: "subdir1", Artefact: "c1"},
					Target: target.Target{Dir: "subdir2", Artefact: "d4"},
				},
			},
		},
	} {
		t.Run(test.Name, func(u *testing.T) {
			pruned := pruneLinks(test.Links, test.Configs)
			assert.Equal(u, test.Expected, pruned)
		})
	}
}
