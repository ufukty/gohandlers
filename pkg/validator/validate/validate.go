// Package validate adds convenience functions for validating values in
// common types. Validation methods are straigtforward and don't aim for
// being the best fit for every use.
package validate

import (
	"fmt"
	"regexp"
	"time"
)

var (
	ErrPattern = fmt.Errorf("content")
	ErrLong    = fmt.Errorf("too long")
	ErrShort   = fmt.Errorf("too short")
	ErrEmpty   = fmt.Errorf("empty")
)

func String(s string, min, max int, pattern *regexp.Regexp) error {
	if l := len(s); l == 0 {
		return ErrEmpty
	} else if l < min {
		return ErrShort
	} else if l > max {
		return ErrLong
	} else if pattern != nil && !pattern.MatchString(string(s)) {
		return ErrPattern
	}
	return nil
}

func Time(t, min, max time.Time) error {
	if t.After(max) {
		return fmt.Errorf("after valid period")
	}
	if t.Before(min) {
		return fmt.Errorf("before valid period")
	}
	return nil
}
