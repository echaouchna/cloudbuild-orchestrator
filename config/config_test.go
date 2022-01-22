package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestFilter(t *testing.T) {
	tcs := []struct {
		name     string
		included []string
		excluded []string
		steps    []Step
		expected []Step
	}{
		{
			name:     "nothing",
			included: []string{},
			excluded: []string{},
			steps:    []Step{},
			expected: []Step{},
		},
		{
			name:     "simple include",
			included: []string{"example"},
			excluded: []string{},
			steps: []Step{
				{
					Name: "a",
					Tags: "example",
				},
			},
			expected: []Step{
				{
					Name: "a",
					Tags: "example",
				},
			},
		},
		{
			name:     "simple include no match",
			included: []string{"exampl"},
			excluded: []string{},
			steps: []Step{
				{
					Name: "a",
					Tags: "example",
				},
			},
			expected: []Step{},
		},
		{
			name:     "simple include wildcard",
			included: []string{"exampl*"},
			excluded: []string{},
			steps: []Step{
				{
					Name: "a",
					Tags: "example",
				},
			},
			expected: []Step{
				{
					Name: "a",
					Tags: "example",
				},
			},
		},
		{
			name:     "simple include wildcard no match",
			included: []string{"xampl*"},
			excluded: []string{},
			steps: []Step{
				{
					Name: "a",
					Tags: "example",
				},
			},
			expected: []Step{},
		},
		{
			name:     "multiple included tags no match",
			included: []string{"tagA", "tagB", "tagC"},
			excluded: []string{},
			steps: []Step{
				{
					Name: "a",
					Tags: "tagB",
				},
			},
			expected: []Step{
				{
					Name: "a",
					Tags: "tagB",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			c := &Config{Steps: tc.steps}
			got := c.Filter(tc.included, tc.excluded)
			if d := cmp.Diff(
				got.Steps,
				tc.expected,
				cmpopts.IgnoreFields(Step{}, "DependsOn", "Description", "Manual", "ProjectId", "Status", "Trigger", "LogUrl"),
				cmpopts.SortSlices(func(a Step, b Step) bool { return a.Name < b.Name })); d != "" {
				t.Errorf("got %v, want %v", got, tc.expected)
			}
		})
	}

}
