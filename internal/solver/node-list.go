package solver

import "ponglehub.co.uk/tools/mudly/internal/runner"

type nodeListElement struct {
	node     *runner.Node
	config   string
	artefact string
	step     int
}

type NodeList struct {
	list []nodeListElement
}

func (n *NodeList) AddNode(config string, artefact string, node *runner.Node) {
	idx := 0

	latest := n.getLastElement(config, artefact)
	if latest != nil {
		idx = latest.step + 1
		node.DependsOn = append(node.DependsOn, latest.node)
	}

	n.list = append(n.list, nodeListElement{
		node:     node,
		config:   config,
		artefact: artefact,
		step:     idx,
	})
}

func (n *NodeList) GetList() []*runner.Node {
	nodes := []*runner.Node{}

	for id := range n.list {
		nodes = append(nodes, n.list[id].node)
	}

	return nodes
}

func (n *NodeList) getLastElement(config string, artefact string) *nodeListElement {
	idx := -1
	var latest *nodeListElement

	for id, node := range n.list {
		if node.config == config && node.artefact == artefact && node.step > idx {
			idx = node.step
			latest = &n.list[id]
		}
	}

	return latest
}

func (n *NodeList) getFirstElement(config string, artefact string) *nodeListElement {
	for id, node := range n.list {
		if node.config == config && node.artefact == artefact && node.step == 0 {
			return &n.list[id]
		}
	}

	return nil
}

func (n *NodeList) GetLast(config string, artefact string) *runner.Node {
	latest := n.getLastElement(config, artefact)

	if latest == nil {
		return nil
	}

	return latest.node
}
