package test

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
)

// MatchText matches multi-line text ignoring blank lines,
// and leading/trailing space.
// On failure gives a diff-style message useful for long strings.
func MatchLines(expected string) types.GomegaMatcher {
	return &lineMatcher{expected: expected}
}

type lineMatcher struct {
	expected interface{}
	diff     string
}

func (m *lineMatcher) Match(actual interface{}) (success bool, err error) {
	m.diff = cmp.Diff(normalize(m.expected.(string)), normalize(actual.(string)))
	return m.diff == "", nil
}

func (m *lineMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Unexpected diff (-expected, +actual):\n%s\n====\nActual value:\n%s\n", m.diff, actual)
}

func (m *lineMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return ("Expected differences but none found.")
}

func normalize(in string) []string {
	out := []string{}
	for _, line := range strings.Split(in, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			out = append(out, strings.TrimSpace(line))
		}
	}
	return out
}

// EqualDiff is like Equal but gives cmp.Diff style output instead of expected/actual.
func EqualDiff(expect interface{}) types.GomegaMatcher {
	return &equalDiff{matchers.EqualMatcher{Expected: expect}}
}

type equalDiff struct{ matchers.EqualMatcher }

func (m *equalDiff) FailureMessage(actual interface{}) (message string) {
	return "Unexpected diff (-expected, +actual):\n" + cmp.Diff(m.EqualMatcher.Expected, actual)
}
