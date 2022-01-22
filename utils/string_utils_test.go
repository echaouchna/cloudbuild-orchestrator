package utils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestRemoveEmptyStrings(t *testing.T) {
	tcs := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "one-empty",
			input:    []string{""},
			expected: []string{},
		},
		{
			name:     "trailing-spaces",
			input:    []string{"aaa   ", "    bbb", "   ccc  "},
			expected: []string{"aaa", "bbb", "ccc"},
		},
		{
			name:     "empty-element",
			input:    []string{"aaa", "", "bbb"},
			expected: []string{"bbb", "aaa"},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := RemoveEmptyStrings(tc.input)

			if len(got) == len(tc.expected) && len(got) == 0 {
				return
			}

			if d := cmp.Diff(
				got,
				tc.expected,
				cmpopts.SortSlices(func(a string, b string) bool { return a < b })); d != "" {
				t.Errorf("got %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestWildCardToRegexp(t *testing.T) {
	tcs := []struct {
		name               string
		input              string
		expectedString     string
		expectedIsWildCard bool
	}{
		{
			name:               "empty",
			input:              "",
			expectedString:     "",
			expectedIsWildCard: false,
		},
		{
			name:               "non-wildcard",
			input:              "aaa",
			expectedString:     "",
			expectedIsWildCard: false,
		},
		{
			name:               "simple-wildcard",
			input:              "test*",
			expectedString:     "^test.*$",
			expectedIsWildCard: true,
		},
		{
			name:               "multiple-wildcard",
			input:              "*test*-*",
			expectedString:     "^.*test.*-.*$",
			expectedIsWildCard: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			isWildcard, actualString := WildCardToRegexp(tc.input)

			if isWildcard != tc.expectedIsWildCard {
				t.Errorf("is wildcar got %v, want %v", actualString, tc.expectedString)
			}

			if isWildcard && actualString != tc.expectedString {
				t.Errorf("wildcard to regex got %v, want %v", actualString, tc.expectedString)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tcs := []struct {
		name     string
		slice    []string
		element  string
		expected bool
	}{
		{
			name:     "empty",
			slice:    []string{},
			element:  "",
			expected: false,
		},
		{
			name:     "empty-slice",
			slice:    []string{},
			element:  "aaa",
			expected: false,
		},
		{
			name:     "empty-element",
			slice:    []string{"aaa", "bbb", "ccc"},
			element:  "",
			expected: false,
		},
		{
			name:     "no-match",
			slice:    []string{"aaa", "bbb", "ccc"},
			element:  "ddd",
			expected: false,
		},
		{
			name:     "match",
			slice:    []string{"aaa", "bbb", "ccc"},
			element:  "bbb",
			expected: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := Contains(tc.slice, tc.element)

			if got != tc.expected {
				t.Errorf("got %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestMatchAtLeastOne(t *testing.T) {
	tcs := []struct {
		name     string
		values   []string
		input    string
		expected bool
	}{
		{
			name:     "empty",
			values:   []string{},
			input:    "",
			expected: false,
		},
		{
			name:     "empty-values",
			values:   []string{},
			input:    "aaa",
			expected: false,
		},
		{
			name:     "no-match",
			values:   []string{"testA"},
			input:    "test",
			expected: false,
		},
		{
			name:     "exact-match",
			values:   []string{"testA"},
			input:    "testA",
			expected: true,
		},
		{
			name:     "one-of",
			values:   []string{"testA", "testB", "testC"},
			input:    "testB",
			expected: true,
		},
		{
			name:     "one-of-regex",
			values:   []string{"testA", "test*", "testC"},
			input:    "testB",
			expected: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := MatchAtLeastOne(tc.values, tc.input)

			if got != tc.expected {
				t.Errorf("got %v, want %v", got, tc.expected)
			}
		})
	}
}
