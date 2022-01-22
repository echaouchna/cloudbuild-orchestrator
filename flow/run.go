package flow

import (
	"cork/cmd"
	"cork/config"
	"cork/dag"
	"cork/gcp"
	"cork/utils"
	"errors"
	"fmt"
	"sync"
)

type executionContext struct {
	lock     sync.Mutex
	conf     *config.Config
	exactRef string
	options  cmd.Options
	triggers map[string]*gcp.BuildTrigger
	dag      *dag.Dag
}

func listUniqueProjects(d *dag.Dag) []string {
	uniqueProjectIDs := []string{}
	for _, node := range d.Nodes {
		projectID := node.Task.(config.Step).ProjectId
		if !utils.Contains(uniqueProjectIDs, projectID) {
			uniqueProjectIDs = append(uniqueProjectIDs, projectID)
		}
	}
	return uniqueProjectIDs
}

func listTriggers(d *dag.Dag) map[string]*gcp.BuildTrigger {
	triggers := map[string]*gcp.BuildTrigger{}
	uniqueProjects := listUniqueProjects(d)
	for _, project := range uniqueProjects {
		for k, v := range gcp.ListTriggers(project) {
			triggers[k] = v
		}
	}
	return triggers
}

func waitForDepBuilds(ctx *executionContext, step config.Step, triggerName string) error {
	if step.Manual {
		for _, dep := range step.DependsOn {
			if ctx.dag.Nodes[dep].Task.(config.Step).Status != gcp.SUCCESS {
				message := step.Name + " depends on " + dep + " that has status " + ctx.dag.Nodes[dep].Task.(config.Step).Status
				flowLog(Log{Message: message, Progress: SKIP})
				return nil
			}
			ynResponse := waitForInput(WaitInput{
				Trigger: triggerName,
				Message: fmt.Sprintf("Please validate %s to continue", dep),
				LogUrl:  ctx.dag.Nodes[dep].Task.(config.Step).LogUrl,
			})

			if !ynResponse {
				message := triggerName + " cancelled by user"
				flowLog(Log{Message: message, Progress: SKIP})
				return errors.New(message)
			}
		}
	}
	return nil
}

func handleTrigger(node *dag.Node, ctx *executionContext) error {
	step := node.Task.(config.Step)
	defer func() {
		node.Task = step
	}()
	triggerFullName := step.ProjectId + "/" + step.Trigger
	func() {
		ctx.lock.Lock()
		defer ctx.lock.Unlock()
		step.Status = SKIP
	}()
	buildTrigger := ctx.triggers[triggerFullName]
	ref := getRef(ctx.options.Reference, ctx.exactRef)
	if buildTrigger == nil {
		message := ctx.conf.Name + " no trigger matching " + triggerFullName + " found"
		flowLog(Log{Message: message, Progress: SKIP})
		return errors.New(message)
	}
	triggerName := ctx.conf.Name + "/" + buildTrigger.Name
	err := waitForDepBuilds(ctx, step, triggerName)
	if err != nil {
		return err
	}
	flowLog(Log{Trigger: triggerName, Message: "started", Progress: gcp.RUNNING})
	build, err := gcp.TriggerCloudBuild(
		step.ProjectId,
		buildTrigger.Id,
		getSourceRepo(ref),
	)
	if err != nil {
		flowLog(Log{
			Trigger:  triggerName,
			Message:  err.Error(),
			Progress: gcp.FAILURE,
		})
		return err
	}

	func() {
		ctx.lock.Lock()
		defer ctx.lock.Unlock()
		setExactRef(&ctx.exactRef, build)
	}()

	flowLog(Log{
		Trigger:  triggerName,
		Message:  "triggered",
		LogUrl:   build.LogURL,
		Progress: gcp.RUNNING,
	})

	status := waitForBuild(step.ProjectId, build.ID)
	func() {
		ctx.lock.Lock()
		defer ctx.lock.Unlock()
		step.Status = status
		step.LogUrl = build.LogURL
	}()

	switch status {
	case gcp.SUCCESS:
		flowLog(Log{
			Trigger:  triggerName,
			Message:  "finished",
			LogUrl:   build.LogURL,
			Progress: status,
		})
	case gcp.FAILURE, gcp.CANCELLED:
		flowLog(Log{
			Trigger:  triggerName,
			Message:  status,
			LogUrl:   build.LogURL,
			Progress: status,
		})
		return errors.New("build failed")
	default:
		return errors.New("Unknown status " + status)
	}
	return nil
}

func runJob(jobs chan *dag.Node, results chan error, ctx *executionContext) {
	for j := range jobs {
		err := handleTrigger(j, ctx)
		results <- err
	}
}

func initJobs(jobs chan *dag.Node, results chan error, ctx *executionContext) {
	for w := 0; w < ctx.options.NumParallelJobs; w++ {
		go runJob(jobs, results, ctx)
	}
}

func startStep(jobs chan *dag.Node, node *dag.Node) {
	step := node.Task.(config.Step)
	step.Status = gcp.RUNNING
	node.Task = step
	jobs <- node
}

func waitForResults(options cmd.Options, jobsNumber int, d *dag.Dag, jobs chan *dag.Node, results chan error) {
	for i := 0; i < jobsNumber; i++ {
		result := <-results
		if result != nil && !options.NoFastFailing {
			fmt.Println(result.Error())
			fmt.Println("Fast failing")
			return
		}

		schedulableStepKeys, err := d.GetNodesToSchedule()
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		for _, stepKey := range schedulableStepKeys {
			startStep(jobs, d.Nodes[stepKey])
		}
	}
}

func run(d *dag.Dag, conf *config.Config, options cmd.Options) {
	triggers := listTriggers(d)

	jobs := make(chan *dag.Node, len(d.Nodes))
	defer close(jobs)

	results := make(chan error, len(d.Nodes))
	defer close(results)

	ctx := &executionContext{
		options:  options,
		conf:     conf,
		triggers: triggers,
		dag:      d,
	}

	initJobs(jobs, results, ctx)

	schedulableStepKeys, err := d.GetNodesToSchedule()
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, stepKey := range schedulableStepKeys {
		startStep(jobs, d.Nodes[stepKey])
	}

	waitForResults(options, len(d.Nodes), d, jobs, results)
}
