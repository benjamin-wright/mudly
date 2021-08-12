package runner_test

import (
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"ponglehub.co.uk/tools/mudly/internal/runner"
)

type callArgs struct {
	dir      string
	artefact string
	env      map[string]string
}

type callStack struct {
	calls []callArgs
	mu    *sync.Mutex
}

func (c *callStack) addCall(args callArgs) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.calls = append(c.calls, args)
}

type mockRunnable struct {
	result  runner.CommandResult
	stack   *callStack
	timeout int
}

func (n *mockRunnable) Run(dir string, artefact string, env map[string]string) runner.CommandResult {
	time.Sleep(time.Millisecond * time.Duration(n.timeout))
	n.stack.addCall(callArgs{dir: dir, artefact: artefact, env: env})
	return n.result
}

func (n *mockRunnable) String() string { return "" }

func convert(nodesString string) ([]*runner.Node, *callStack) {
	nodes := []*runner.Node{}
	stack := callStack{mu: &sync.Mutex{}}

	for _, segment := range strings.Split(nodesString, ",") {
		parts := strings.Split(segment, ":")
		timeout := 1

		if strings.Contains(parts[0], "(") {
			startIndex := strings.Index(parts[0], "(")
			endIndex := strings.Index(parts[0], ")")
			timeout, _ = strconv.Atoi(parts[0][startIndex+1 : endIndex])
			parts[0] = parts[0][:startIndex]
		}

		node := runner.Node{
			Artefact: parts[0],
			Step: &mockRunnable{
				result:  runner.COMMAND_SUCCESS,
				stack:   &stack,
				timeout: timeout,
			},
			State: runner.STATE_PENDING,
		}

		nodes = append(nodes, &node)
	}

	for index, segment := range strings.Split(nodesString, ",") {
		parts := strings.Split(segment, ":")

		if len(parts) != 2 {
			continue
		}

		for _, extra := range strings.Split(parts[1], "+") {
			for linkIndex, node := range nodes {
				if node.Artefact == extra {
					nodes[index].DependsOn = append(nodes[index].DependsOn, nodes[linkIndex])
				}
			}
		}
	}

	return nodes, &stack
}

func TestRun(t *testing.T) {
	for _, test := range []struct {
		Name     string
		Nodes    string
		Expected string
	}{
		{
			Name:     "test",
			Nodes:    "A",
			Expected: "A",
		},
		{
			Name:     "test",
			Nodes:    "A:B,B:C,C",
			Expected: "C,B,A",
		},
		{
			Name:     "test",
			Nodes:    "A,B:A,C:B",
			Expected: "A,B,C",
		},
		{
			Name:     "test",
			Nodes:    "A(10),B(1),C:A",
			Expected: "B,A,C",
		},
		{
			Name:     "test",
			Nodes:    "A(10),B(1),C:A,D:C+E,E:B",
			Expected: "B,E,A,C,D",
		},
	} {
		t.Run(test.Name, func(u *testing.T) {
			nodes, stack := convert(test.Nodes)
			err := runner.Run(nodes)

			assert.NoError(u, err)

			order := []string{}
			for _, call := range stack.calls {
				order = append(order, call.artefact)
			}

			assert.Equal(u, test.Expected, strings.Join(order, ","))
		})
	}
}
