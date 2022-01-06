package validate

import (
	"cork/config"
	"errors"
	"fmt"
	"strings"
)

type ConfigWarning struct {
	Message string
	Type    string
}

func formatErr(groupName string, err error) string {
	lines := strings.Split(err.Error(), "\n")
	indented := make([]string, len(lines))

	for i, l := range lines {
		indented[i] = "\t" + l
	}

	return fmt.Sprintf("invalid %s:\n%s\n", groupName, strings.Join(indented, "\n"))
}

func compositeErr(errorMessages []string) error {
	if len(errorMessages) == 0 {
		return nil
	}

	return errors.New(strings.Join(errorMessages, "\n"))
}

func ValidateConfig(c config.Config) ([]ConfigWarning, []string) {
	warnings := []ConfigWarning{}
	errorMessages := []string{}
	stepsWarnings, stepsErr := validateSteps(c)
	if stepsErr != nil {
		errorMessages = append(errorMessages, formatErr("steps", stepsErr))
	}
	warnings = append(warnings, stepsWarnings...)
	return warnings, errorMessages
}

func validateSteps(c config.Config) ([]ConfigWarning, error) {
	warnings := []ConfigWarning{}
	errorMessages := []string{}
	if len(c.Steps) == 0 {
		message := "no steps defined"
		warnings = append(warnings, ConfigWarning{Type: "steps", Message: message})
		errorMessages = append(errorMessages, message)
	}

	for _, step := range c.Steps {
		if len(step.Parallel) > 0 {
			parallelWarnings, parallelErr := validateParallel(step)
			if parallelErr != nil {
				errorMessages = append(errorMessages, formatErr("parallel steps", parallelErr))
			}
			warnings = append(warnings, parallelWarnings...)
		} else {
			if step.Name == "" {
				errorMessages = append(errorMessages, "step missing name")
			}
			if step.ProjectId == "" {
				errorMessages = append(errorMessages, "step missing project-id")
			}
			if step.Trigger == "" && len(step.Parallel) == 0 {
				errorMessages = append(errorMessages, "step missing trigger")
			}
		}
	}
	return warnings, compositeErr(errorMessages)
}

func validateParallel(parallelStep config.Step) ([]ConfigWarning, error) {
	warnings := []ConfigWarning{}
	errorMessages := []string{}

	if parallelStep.Name != "" {
		warnings = append(warnings, ConfigWarning{Type: "step", Message: "Name not needed for parallel steps"})
	}

	if parallelStep.ProjectId != "" {
		warnings = append(warnings, ConfigWarning{Type: "step", Message: "Project ID not needed for parallel steps"})
	}

	if parallelStep.Trigger != "" {
		errorMessages = append(errorMessages, "parallel steps cannot be used when a trigger is defined")
	}

	for _, step := range parallelStep.Parallel {
		if step.Name == "" {
			errorMessages = append(errorMessages, "step missing name")
		}
		if step.ProjectId == "" {
			errorMessages = append(errorMessages, "step missing project-id")
		}
		if step.Trigger == "" {
			errorMessages = append(errorMessages, "step missing trigger")
		}
	}
	return warnings, compositeErr(errorMessages)
}
