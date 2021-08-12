package runner

type NodeState int

const (
	STATE_PENDING NodeState = iota
	STATE_RUNNING
	STATE_ERROR
	STATE_SKIPPED
	STATE_COMPLETE
)

type CommandResult int

const (
	COMMAND_SUCCESS CommandResult = iota
	COMMAND_ERROR
	COMMAND_SKIPPED
	COMMAND_SKIP_ARTEFACT
)

type Runnable interface {
	Run(dir string, artefact string, env map[string]string) CommandResult
	String() string
}

type Node struct {
	SharedEnv map[string]string
	Path      string
	Artefact  string
	Step      Runnable
	State     NodeState
	DependsOn []*Node
}
