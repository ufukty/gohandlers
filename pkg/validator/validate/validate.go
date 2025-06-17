// Package validate adds convenience functions for validating values in
// common types. Validation methods are straigtforward and don't aim for
// being the best fit for every use.
package validate

import (
	"regexp"
	"time"
)

var (
	IssueEmpty   = "empty"
	IssueLong    = "too long"
	IssuePattern = "pattern"
	IssueShort   = "too short"
)

func String(s string, min, max int, pattern *regexp.Regexp) any {
	if l := len(s); l == 0 {
		return IssueEmpty
	} else if l < min {
		return IssueShort
	} else if l > max {
		return IssueLong
	} else if pattern != nil && !pattern.MatchString(string(s)) {
		return IssuePattern
	}
	return nil
}

var (
	IssueAfterValidPeriod  = "after valid period"
	IssueBeforeValidPeriod = "before valid period"
)

func Time(t, min, max time.Time) any {
	if t.After(max) {
		return IssueAfterValidPeriod
	}
	if t.Before(min) {
		return IssueBeforeValidPeriod
	}
	return nil
}
