package dag

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type testTask struct {
	name   string
	status string
	deps   []string
}

type testTaskList []testTask

func (ttl testTaskList) Items() []Task {
	tasks := []Task{}
	for _, t := range ttl {
		tasks = append(tasks, Task(t))
	}
	return tasks
}

func (tt testTask) GetKey() string {
	return tt.name
}

func (tt testTask) HasFinished() bool {
	return tt.status == "done"
}

func (tt testTask) IsSuccessful() bool {
	return tt.status == "done"
}

func (tt testTask) HasStarted() bool {
	return tt.status != ""
}

func buildTestDag(t *testing.T) *Dag {
	//  u     a   x
	//  |    / \ /
	//  |   |   y
	//  |   | / |
	//  |   b   |
	//   \ /    z
	//    v
	//    |
	//    w
	t.Helper()
	tasks := []testTask{
		{
			name: "a",
		}, {
			name: "u",
		}, {
			name: "x",
		}, {
			name: "w",
			deps: []string{"v"},
		}, {
			name: "v",
			deps: []string{"b", "u"},
		}, {
			name: "y",
			deps: []string{"a", "x"},
		}, {
			name: "b",
			deps: []string{"a", "y"},
		}, {
			name: "z",
			deps: []string{"y"},
		},
	}
	links := map[string][]string{}
	for _, task := range tasks {
		if len(task.deps) > 0 {
			links[task.name] = task.deps
		}
	}
	d, err := BuildDag(testTaskList(tasks), links)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestBuildDag(t *testing.T) {
	d := buildTestDag(t)
	expectedDag := &Dag{
		Nodes: map[string]*Node{
			"a": {
				Task: Task(testTask{
					name: "a",
				}),
			},
			"b": {
				Task: Task(testTask{
					name: "b",
				}),
			},
			"u": {
				Task: Task(testTask{
					name: "u",
				}),
			},
			"v": {
				Task: Task(testTask{
					name: "v",
				}),
			},
			"w": {
				Task: Task(testTask{
					name: "w",
				}),
			},
			"x": {
				Task: Task(testTask{
					name: "x",
				}),
			},
			"y": {
				Task: Task(testTask{
					name: "y",
				}),
			},
			"z": {
				Task: Task(testTask{
					name: "z",
				}),
			},
		},
	}
	expectedDag.Nodes["a"].Next = []*Node{expectedDag.Nodes["b"], expectedDag.Nodes["y"]}
	expectedDag.Nodes["b"].Next = []*Node{expectedDag.Nodes["v"]}
	expectedDag.Nodes["u"].Next = []*Node{expectedDag.Nodes["v"]}
	expectedDag.Nodes["v"].Next = []*Node{expectedDag.Nodes["w"]}
	expectedDag.Nodes["x"].Next = []*Node{expectedDag.Nodes["y"]}
	expectedDag.Nodes["y"].Next = []*Node{expectedDag.Nodes["b"], expectedDag.Nodes["z"]}
	expectedDag.Nodes["b"].Prev = []*Node{expectedDag.Nodes["a"], expectedDag.Nodes["y"]}
	expectedDag.Nodes["v"].Prev = []*Node{expectedDag.Nodes["b"], expectedDag.Nodes["u"]}
	expectedDag.Nodes["w"].Prev = []*Node{expectedDag.Nodes["v"]}
	expectedDag.Nodes["y"].Prev = []*Node{expectedDag.Nodes["a"], expectedDag.Nodes["x"]}
	expectedDag.Nodes["z"].Prev = []*Node{expectedDag.Nodes["y"]}
	assertSameDag(t, expectedDag, d)
}

func TestBuildDagInvalid(t *testing.T) {
	tcs := []struct {
		name string
		spec []testTask
		err  string
	}{
		{
			name: "invalid link",
			spec: []testTask{
				{
					name: "a",
					deps: []string{"w"},
				}, {
					name: "b",
				},
			},
			err: "couldn't add link",
		},
		{
			name: "duplicate task",
			spec: []testTask{
				{
					name: "a",
				}, {
					name: "b",
				}, {
					name: "a",
				},
			},
			err: "already present in the Dag: duplicate task",
		},
		{
			name: "self cycle",
			spec: []testTask{
				{
					name: "a",
					deps: []string{"b"},
				}, {
					name: "b",
					deps: []string{"b"},
				}, {
					name: "c",
				},
			},
			err: "depends on itself",
		},
		{
			name: "cycle",
			spec: []testTask{
				{
					name: "a",
					deps: []string{"w"},
				}, {
					name: "b",
				}, {
					name: "w",
					deps: []string{"b", "y"},
				}, {
					name: "x",
					deps: []string{"a"},
				}, {
					name: "y",
					deps: []string{"a", "x"},
				}, {
					name: "z",
					deps: []string{"x"},
				},
			},
			err: "cycle detected: cycle detected:",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			links := map[string][]string{}
			for _, task := range tc.spec {
				if len(task.deps) > 0 {
					links[task.name] = task.deps
				}
			}
			_, err := BuildDag(testTaskList(tc.spec), links)
			if err == nil || !strings.Contains(err.Error(), tc.err) {
				t.Errorf("expected to see an error for invalid dag %v but had none", tc.spec)
			}
		})
	}
}

func TestGetSchedulable(t *testing.T) {
	tcs := []struct {
		name          string
		finished      []string
		expectedTasks []string
	}{
		{
			name:          "nothing-done",
			finished:      []string{},
			expectedTasks: []string{"a", "u", "x"},
		}, {
			name:          "a-done",
			finished:      []string{"a"},
			expectedTasks: []string{"u", "x"},
		}, {
			name:          "x-done",
			finished:      []string{"x"},
			expectedTasks: []string{"a", "u"},
		}, {
			name:          "u-done",
			finished:      []string{"u"},
			expectedTasks: []string{"a", "x"},
		}, {
			name:          "a-x-done",
			finished:      []string{"a", "x"},
			expectedTasks: []string{"u", "y"},
		}, {
			name:          "a-u-done",
			finished:      []string{"a", "u"},
			expectedTasks: []string{"x"},
		}, {
			name:          "x-u-done",
			finished:      []string{"x", "u"},
			expectedTasks: []string{"a"},
		}, {
			name:          "a-x-u-done",
			finished:      []string{"a", "x", "u"},
			expectedTasks: []string{"y"},
		}, {
			name:          "a-x-y-done",
			finished:      []string{"a", "x", "y"},
			expectedTasks: []string{"b", "u", "z"},
		}, {
			name:          "a-x-y-b-done",
			finished:      []string{"a", "x", "y", "b"},
			expectedTasks: []string{"u", "z"},
		}, {
			name:          "a-x-u-y-done",
			finished:      []string{"a", "x", "u", "y"},
			expectedTasks: []string{"b", "z"},
		}, {
			name:          "a-x-y-z-done",
			finished:      []string{"a", "x", "z", "y"},
			expectedTasks: []string{"b", "u"},
		}, {
			name:          "a-x-u-y-b-done",
			finished:      []string{"a", "x", "u", "y", "b"},
			expectedTasks: []string{"v", "z"},
		}, {
			name:          "a-x-u-y-z-done",
			finished:      []string{"a", "x", "u", "y", "z"},
			expectedTasks: []string{"b"},
		}, {
			name:          "a-x-u-y-b-v-done",
			finished:      []string{"a", "x", "u", "y", "b", "v"},
			expectedTasks: []string{"z", "w"},
		}, {
			name:          "a-x-u-y-b-z-done",
			finished:      []string{"a", "x", "u", "y", "b", "z"},
			expectedTasks: []string{"v"},
		}, {
			name:          "a-x-u-y-b-z-v-done",
			finished:      []string{"a", "x", "u", "y", "b", "z", "v"},
			expectedTasks: []string{"w"},
		},
	}
	for _, tc := range tcs {
		d := buildTestDag(t)
		t.Run(tc.name, func(t *testing.T) {
			for _, task := range tc.finished {
				taskT := d.Nodes[task].Task.(testTask)
				taskT.status = "done"
				d.Nodes[task].Task = taskT
			}
			tasks, err := d.GetNodesToSchedule()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if d := cmp.Diff(tasks, tc.expectedTasks, cmpopts.IgnoreFields(testTask{}, "deps", "status"), cmpopts.SortSlices(func(a string, b string) bool { return a < b })); d != "" {
				t.Errorf("expected that with %v done, %v would be ready to schedule but was different: %s", tc.finished, tc.expectedTasks, PrintWantGot(t, d))
			}
		})
	}
}

func TestGetSchedulableInvalid(t *testing.T) {
	tcs := []struct {
		name     string
		finished []string
	}{
		{
			name:     "only-z",
			finished: []string{"z"},
		}, {
			name:     "only-y",
			finished: []string{"y"},
		}, {
			name:     "only-w",
			finished: []string{"w"},
		}, {
			name:     "only-y-x",
			finished: []string{"y", "x"},
		}, {
			name:     "only-y-w",
			finished: []string{"y", "w"},
		}, {
			name:     "only-x-w",
			finished: []string{"x", "w"},
		},
	}
	for _, tc := range tcs {
		d := buildTestDag(t)
		t.Run(tc.name, func(t *testing.T) {
			for _, task := range tc.finished {
				taskT := d.Nodes[task].Task.(testTask)
				taskT.status = "done"
				d.Nodes[task].Task = taskT
			}
			_, err := d.GetNodesToSchedule()
			if err == nil {
				t.Fatalf("Expected error for invalid done tasks %v but got none", tc.finished)
			}
		})
	}
}

func PrintWantGot(t *testing.T, diff string) string {
	t.Helper()
	return fmt.Sprintf("(-want, +got): %s", diff)
}

func compareStringSlices(t *testing.T, expected, actual []string) error {
	t.Helper()
	if d := cmp.Diff(expected, actual,
		cmpopts.SortSlices(func(a string, b string) bool { return a < b }),
	); d != "" {
		return fmt.Errorf("Dags contain different nodes: %s", PrintWantGot(t, d))
	}
	return nil
}

func sameNodes(t *testing.T, expected, actual []*Node) error {
	expectedNames, actualNames := []string{}, []string{}
	for _, n := range expected {
		expectedNames = append(expectedNames, n.Task.GetKey())
	}
	for _, n := range actual {
		actualNames = append(actualNames, n.Task.GetKey())
	}

	return compareStringSlices(t, expectedNames, actualNames)
}

func sameDagNodes(t *testing.T, expected, actual *Dag) error {
	expectedKeys, actualKeys := []string{}, []string{}

	for k := range expected.Nodes {
		expectedKeys = append(expectedKeys, k)
	}
	for k := range actual.Nodes {
		actualKeys = append(actualKeys, k)
	}

	return compareStringSlices(t, expectedKeys, actualKeys)
}

func assertSameDag(t *testing.T, expected, actual *Dag) {
	t.Helper()

	err := sameDagNodes(t, expected, actual)
	if err != nil {
		t.Errorf("Dags contain different nodes: %s", err)
		return
	}

	for k, actualNode := range actual.Nodes {
		expectedNode := expected.Nodes[k]

		err := sameNodes(t, actualNode.Prev, expectedNode.Prev)
		if err != nil {
			t.Errorf("The %s nodes in the dag have different previous nodes: %v", k, err)
		}
		err = sameNodes(t, actualNode.Next, expectedNode.Next)
		if err != nil {
			t.Errorf("The %s nodes in the dag have different next nodes: %v", k, err)
		}
	}
}
