package runner

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"ponglehub.co.uk/tools/mudly/internal/utils"
)

type runResult struct {
	node   *Node
	result CommandResult
}

func Run(nodes []*Node) (err error) {
	numRunning := 0
	outputChan := make(chan runResult, 10)

	for {
		pending := getRunnableNodes(nodes)

		for _, node := range pending {
			numRunning += 1
			node.State = STATE_RUNNING
			go runNode(node, outputChan)
		}

		if numRunning == 0 {
			return nil
		}

		result := <-outputChan

		switch result.result {
		case COMMAND_ERROR:
			return fmt.Errorf("error running step %s:%s", result.node.Artefact, result.node.Step)
		case COMMAND_SKIP_ARTEFACT:
			for _, node := range nodes {
				if node.Path == result.node.Path && node.Artefact == result.node.Artefact {
					node.State = STATE_SKIPPED
				}
			}
		}

		numRunning -= 1
	}
}

func getRunnableNodes(nodes []*Node) []*Node {
	runnables := []*Node{}

	for _, node := range nodes {
		runnable := node.State == STATE_PENDING

		for _, dep := range node.DependsOn {
			if dep.State != STATE_COMPLETE && dep.State != STATE_SKIPPED {
				runnable = false
			}
		}

		if runnable {
			runnables = append(runnables, node)
		}
	}

	return runnables
}

func runNode(node *Node, outputChan chan<- runResult) {
	logrus.Infof("{%s} %s[%s]: STARTED", node.Path, node.Artefact, node.Step)

	merged := utils.MergeMaps(
		node.SharedEnv,
		map[string]string{
			"MUDLY_PWD": os.Getenv("PWD"),
		},
	)

	result := node.Step.Run(node.Path, node.Artefact, merged)

	switch result {
	case COMMAND_SUCCESS:
		logrus.Infof("{%s} %s[%s]: DONE", node.Path, node.Artefact, node.Step)
		node.State = STATE_COMPLETE
	case COMMAND_SKIPPED:
		logrus.Infof("{%s} %s[%s]: SKIPPED", node.Path, node.Artefact, node.Step)
		node.State = STATE_SKIPPED
	case COMMAND_SKIP_ARTEFACT:
		logrus.Infof("{%s} %s[%s]: SKIPPED ARTEFACT", node.Path, node.Artefact, node.Step)
		node.State = STATE_SKIPPED
	case COMMAND_ERROR:
		logrus.Infof("{%s} %s[%s]: ERROR", node.Path, node.Artefact, node.Step)
		node.State = STATE_ERROR
	}

	outputChan <- runResult{
		node:   node,
		result: result,
	}
}
