// Package validator contains the symbols act as constructors for value
// validators. They are designed to run on app initialization and provide
// caller an interface for constructing validators for all custom types
// with the least lines of code.
package validator

import (
	"regexp"

	"github.com/ufukty/gohandlers/pkg/validator/validate"
)

type Strings struct {
	pattern  *regexp.Regexp
	min, max int
}

func (v Strings) Validate(s string) any {
	return validate.String(s, v.min, v.max, v.pattern)
}

func ForStrings(pattern string, min, max int) Strings {
	return Strings{
		pattern: regexp.MustCompile(pattern),
		min:     min,
		max:     max,
	}
}
