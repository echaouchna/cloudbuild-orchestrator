package flow

import (
	"cork/config"
	"cork/gcp"
	"cork/utils"
	"errors"
	"fmt"

	"github.com/workanator/go-floc/v3"
	runFloc "github.com/workanator/go-floc/v3/run"
)

func listUniqueProjects(config config.Config) []string {
	uniqueProjects := []string{}
	for _, step := range config.Steps {
		var project string
		if len(step.Parallel) > 0 {
			for _, parallelStep := range step.Parallel {
				project = parallelStep.ProjectId
				if !utils.Contains(uniqueProjects, project) {
					uniqueProjects = append(uniqueProjects, project)
				}
			}
		} else {
			project = step.ProjectId
		}
		if !utils.Contains(uniqueProjects, project) {
			uniqueProjects = append(uniqueProjects, project)
		}
	}
	return uniqueProjects
}

func listTriggers(config config.Config, executionFlowContext *ExecutionFlowContext) {
	uniqueProjects := listUniqueProjects(config)
	for _, project := range uniqueProjects {
		for k, v := range gcp.ListTriggers(project) {
			executionFlowContext.triggers[k] = v
		}
	}
}

func shouldHandleTrigger(options Options, stepType string) bool {
	if (len(options.IncludedTypes) > 0 &&
		!utils.MatchAtLeastOne(options.IncludedTypes, stepType)) ||
		(len(options.ExcludedTypes) > 0 &&
			utils.MatchAtLeastOne(options.ExcludedTypes, stepType)) {
		return false
	}
	return true
}

func returnConditionally(noFastFailing bool, err error) error {
	if noFastFailing {
		return nil
	}
	return err
}

func handleTrigger(executionFlowContext *ExecutionFlowContext, step config.Step) error {
	triggerFullName := step.ProjectId + "/" + step.Trigger
	func() {
		executionFlowContext.lock.Lock()
		defer executionFlowContext.lock.Unlock()
		executionFlowContext.statuses[step.Name] = BuildStatus{
			value:  SKIP,
			logUrl: "",
		}
	}()
	if !shouldHandleTrigger(executionFlowContext.options, step.Type) {
		return nil
	}
	buildTrigger := executionFlowContext.triggers[triggerFullName]
	ref := getRef(executionFlowContext.options.Reference, executionFlowContext.exactRef)
	if buildTrigger == nil {
		message := executionFlowContext.config.Name + " no trigger matching " + triggerFullName + " found"
		flowLog(Log{Message: message, Progress: SKIP})
		return returnConditionally(executionFlowContext.options.NoFastFailing, errors.New(message))
	}
	triggerName := executionFlowContext.config.Name + "/" + buildTrigger.Name
	if step.DependsOn != "" {
		if executionFlowContext.options.NoFastFailing && executionFlowContext.statuses[step.DependsOn].value != gcp.SUCCESS {
			message := step.Name + " depends on " + step.DependsOn + " that has status " + executionFlowContext.statuses[step.DependsOn].value
			flowLog(Log{Message: message, Progress: SKIP})
			return nil
		}
		ynResponse := waitForInput(WaitInput{
			Trigger: triggerName,
			Message: fmt.Sprintf("Please validate %s to continue", step.DependsOn),
			LogUrl:  executionFlowContext.statuses[step.DependsOn].logUrl,
		})

		if !ynResponse {
			message := triggerName + " cancelled by user"
			flowLog(Log{Message: message, Progress: SKIP})
			return returnConditionally(executionFlowContext.options.NoFastFailing, errors.New(message))
		}
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
		return returnConditionally(executionFlowContext.options.NoFastFailing, err)
	}

	func() {
		executionFlowContext.lock.Lock()
		defer executionFlowContext.lock.Unlock()
		setExactRef(&executionFlowContext.exactRef, build)
	}()

	flowLog(Log{
		Trigger:  triggerName,
		Message:  "triggered",
		LogUrl:   build.LogURL,
		Progress: gcp.RUNNING,
	})

	status := waitForBuild(step.ProjectId, build.ID)
	func() {
		executionFlowContext.lock.Lock()
		defer executionFlowContext.lock.Unlock()
		executionFlowContext.statuses[step.Name] = BuildStatus{
			value:  status,
			logUrl: build.LogURL,
		}
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
		return returnConditionally(executionFlowContext.options.NoFastFailing, errors.New("build failed"))
	default:
		return errors.New("Unknown status " + status)
	}
	return nil
}

func buildSequence(conf config.Config) floc.Job {
	jobs := []floc.Job{}

	for _, s := range conf.Steps {
		step := s
		var job floc.Job
		if len(step.Parallel) == 0 {
			job = func(ctx floc.Context, ctrl floc.Control) error {
				executionFlowContext := ctx.Value(1).(*ExecutionFlowContext)
				return handleTrigger(executionFlowContext, step)
			}
		} else {
			parallelJobs := []floc.Job{}
			for _, ps := range step.Parallel {
				parallelStep := ps
				parallelJobs = append(parallelJobs, func(ctx floc.Context, ctrl floc.Control) error {
					executionFlowContext := ctx.Value(1).(*ExecutionFlowContext)
					return handleTrigger(executionFlowContext, config.Step{
						ProjectId: parallelStep.ProjectId,
						Trigger:   parallelStep.Trigger,
						Type:      parallelStep.Type,
						Name:      parallelStep.Name,
						DependsOn: parallelStep.DependsOn,
					})
				})
			}
			job = runFloc.Parallel(parallelJobs...)
		}
		jobs = append(jobs, job)
	}

	return runFloc.Sequence(jobs...)
}

func run(config config.Config, options Options) {
	flowCtx := floc.NewContext()

	executionFlowContext := &ExecutionFlowContext{
		config:   config,
		options:  options,
		triggers: make(map[string]*gcp.BuildTrigger),
		statuses: make(map[string]BuildStatus),
	}

	flowCtx.AddValue(1, executionFlowContext)

	listTriggers(config, executionFlowContext)

	flow := buildSequence(config)

	ctrl := floc.NewControl(flowCtx)

	_, _, err := floc.RunWith(flowCtx, ctrl, flow)
	if err != nil {
		fmt.Println(err)
	}
}
