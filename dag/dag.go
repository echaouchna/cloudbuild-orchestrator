package dag

import (
	"errors"
	"fmt"
	"strings"
)

type (
	Task interface {
		GetKey() string
		HasFinished() bool
		IsSuccessful() bool
		HasStarted() bool
	}

	Node struct {
		Task Task
		Prev []*Node
		Next []*Node
	}

	Dag struct {
		Nodes map[string]*Node
	}

	Tasks interface {
		Items() []Task
	}
)

func newDag() *Dag {
	return &Dag{Nodes: map[string]*Node{}}
}

func (dag *Dag) addTask(t Task) (*Node, error) {
	if _, ok := dag.Nodes[t.GetKey()]; ok {
		return nil, errors.New("duplicate task")
	}
	newNode := &Node{
		Task: t,
	}
	dag.Nodes[t.GetKey()] = newNode
	return newNode, nil
}

func (dag *Dag) addDirectedLink(previousTask string, nextTask string) error {
	prev, ok := dag.Nodes[previousTask]
	if !ok {
		return fmt.Errorf("task %s depends on %s but %s wasn't present in the Dag", nextTask, previousTask, previousTask)
	}
	current := dag.Nodes[nextTask]
	if err := linkTasks(prev, current); err != nil {
		return fmt.Errorf("couldn't create link from %s to %s: %w", prev.Task.GetKey(), current.Task.GetKey(), err)
	}
	return nil
}

func linkTasks(prev *Node, next *Node) error {
	// Check for self cycle
	if prev.Task.GetKey() == next.Task.GetKey() {
		return fmt.Errorf("cycle detected; task %q depends on itself", next.Task.GetKey())
	}
	// Check if we are adding cycles.
	path := []string{next.Task.GetKey(), prev.Task.GetKey()}
	if err := lookForNode(prev.Prev, path, next.Task.GetKey()); err != nil {
		return fmt.Errorf("cycle detected: %w", err)
	}
	next.Prev = append(next.Prev, prev)
	prev.Next = append(prev.Next, next)
	return nil
}

func lookForNode(prev []*Node, path []string, target string) error {
	for _, n := range prev {
		path = append(path, n.Task.GetKey())
		if n.Task.GetKey() == target {
			return fmt.Errorf("cycle detected: %s", path)
		}
		if err := lookForNode(n.Prev, path, target); err != nil {
			return err
		}
	}
	return nil
}

func (dag *Dag) getRoots() []*Node {
	roots := []*Node{}
	for _, n := range dag.Nodes {
		if len(n.Prev) == 0 {
			roots = append(roots, n)
		}
	}
	return roots
}

func BuildDag(tasks Tasks, deps map[string][]string) (*Dag, error) {
	dag := newDag()
	for _, t := range tasks.Items() {
		if _, err := dag.addTask(t); err != nil {
			return nil, fmt.Errorf("task %s is already present in the Dag: %w", t.GetKey(), err)
		}
	}
	for task, taskDeps := range deps {
		for _, previousTask := range taskDeps {
			if err := dag.addDirectedLink(previousTask, task); err != nil {
				return nil, fmt.Errorf("couldn't add link between %s and %s: %w", task, previousTask, err)
			}
		}
	}
	return dag, nil
}

func (dag *Dag) traverseAndAnnotateSchedulable(n *Node, nodeSchedulableMap map[string]bool) error {
	schedulable := true
	if !n.Task.HasFinished() && !n.Task.HasStarted() {
		for _, prev := range n.Prev {
			if _, ok := nodeSchedulableMap[prev.Task.GetKey()]; ok && !dag.Nodes[prev.Task.GetKey()].Task.HasFinished() &&
				!dag.Nodes[prev.Task.GetKey()].Task.IsSuccessful() {
				schedulable = false
				break
			}
		}
	} else if n.Task.HasStarted() && !n.Task.HasFinished() {
		schedulable = false
	} else {
		for _, prev := range n.Prev {
			if !dag.Nodes[prev.Task.GetKey()].Task.HasFinished() {
				return fmt.Errorf("task %s depends on %s but %s hasn't finished yet", n.Task.GetKey(), prev.Task.GetKey(), prev.Task.GetKey())
			}
		}
		schedulable = false
	}
	nodeSchedulableMap[n.Task.GetKey()] = schedulable
	if len(n.Next) > 0 {
		for _, next := range n.Next {
			err := dag.traverseAndAnnotateSchedulable(next, nodeSchedulableMap)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (dag *Dag) GetNodesToSchedule() ([]string, error) {
	roots := dag.getRoots()
	schedulableNodes := []string{}
	nodeSchedulableMap := map[string]bool{}
	for _, root := range roots {
		err := dag.traverseAndAnnotateSchedulable(root, nodeSchedulableMap)
		if err != nil {
			return []string{}, fmt.Errorf("dag status is inconsistent: %w", err)
		}
	}
	for node, schedulable := range nodeSchedulableMap {
		if schedulable {
			schedulableNodes = append(schedulableNodes, node)
		}
	}
	return schedulableNodes, nil
}

func findLinks(node *Node, nextMap map[string][]string) {
	for _, next := range node.Next {
		nextMap[node.Task.GetKey()] = append(nextMap[node.Task.GetKey()], "<"+next.Task.GetKey()+">")
		findLinks(next, nextMap)
	}
}

func (d *Dag) String() string {
	result := "Nodes:\n"
	hasLinks := false
	for k, v := range d.Nodes {
		result += fmt.Sprintf("\t%s\n", k)
		if len(v.Next) > 0 {
			hasLinks = true
		}
	}
	if hasLinks {
		result += "Links:\n"
		nextMap := map[string][]string{}
		for _, node := range d.getRoots() {
			findLinks(node, nextMap)
		}
		for k, v := range nextMap {
			if len(v) > 0 {
				result += fmt.Sprintf("\t%s -> %s\n", k, strings.Join(v, ", "))
			}
		}
	}
	return result
}
