package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReaderPrune(t *testing.T) {
	for _, test := range []struct {
		Name     string
		Lines    []string
		Expected []string
	}{
		{
			Name: "no gaps",
			Lines: []string{
				"ARTEFACT this",
				"  PIPELINE some-pipeline",
			},
			Expected: []string{
				"ARTEFACT this",
				"  PIPELINE some-pipeline",
			},
		},
		{
			Name: "inner gap at same level",
			Lines: []string{
				"ENV SOME=value",
				"",
				"ARTEFACT this",
				"  PIPELINE some-pipeline",
			},
			Expected: []string{
				"ENV SOME=value",
				"ARTEFACT this",
				"  PIPELINE some-pipeline",
			},
		},
		{
			Name: "removes gaps",
			Lines: []string{
				"",
				"ARTEFACT this",
				"  ",
				"  PIPELINE some-pipeline",
			},
			Expected: []string{
				"ARTEFACT this",
				"  PIPELINE some-pipeline",
			},
		},
	} {
		t.Run(test.Name, func(u *testing.T) {
			testReader := reader{
				lines: append([]string{}, test.Lines...),
			}

			testReader.prune()

			assert.Equal(u, test.Expected, testReader.lines)
		})
	}
}
