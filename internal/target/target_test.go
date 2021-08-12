package target_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"ponglehub.co.uk/tools/mudly/internal/target"
)

func TestTargetParsing(t *testing.T) {
	for _, test := range []struct {
		Name     string
		Path     string
		Expected *target.Target
		Error    string
	}{
		{Name: "success", Path: "+target", Expected: &target.Target{Dir: ".", Artefact: "target"}},
		{Name: "with path", Path: "./some/path+target", Expected: &target.Target{Dir: "some/path", Artefact: "target"}},
		{Name: "with special chars", Path: "./some/path+some-target", Expected: &target.Target{Dir: "some/path", Artefact: "some-target"}},
		{Name: "should error if empty", Path: "", Error: "failed to parse target with empty string"},
		{Name: "too many artefacts", Path: "./some/path+target+other", Error: "failed to parse target from path: ./some/path+target+other"},
		{Name: "missing artefact", Path: "./some/path", Error: "failed to parse target from path: ./some/path"},
	} {
		t.Run(test.Name, func(u *testing.T) {
			actual, err := target.ParseTarget(test.Path)

			if test.Expected != nil {
				if actual != nil {
					assert.Equal(u, test.Expected, actual, "Expected parsed target to match")
				} else {
					u.Error("Expected a target")
				}
			}

			if test.Error != "" {
				if err != nil {
					assert.Equal(u, test.Error, err.Error(), "Expected error message to match")
				} else {
					u.Error("Expected an error message")
				}
			}
		})
	}
}
