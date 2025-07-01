package post

import (
	"regexp"
	"strings"
)

var typedecls = regexp.MustCompile(`(?m)$(\n(?://.*\n)*type)`)
var funcdecls = regexp.MustCompile(`(?m)$(\n(?://.*\n)*func)`)
var imports = regexp.MustCompile(`(?m)import \(((?:\s+"[\w]+(?:/[\w.]+)*"\n)*)((?:\s+"[\w.]+(?:/[\w.]+)*"\n)*)\)`)

func Process(s string) string {
	s = typedecls.ReplaceAllString(s, "\n$1")
	s = funcdecls.ReplaceAllString(s, "\n$1")
	s = imports.ReplaceAllString(s, "import ($1\n$2)") // split packages starts with domains
	// 3 lines are for iterators
	s = strings.ReplaceAll(s, "{\"", "{\n\"")
	s = strings.ReplaceAll(s, ", \"", ",\n\"")
	s = strings.ReplaceAll(s, "}\n\t\tfor", ",\n}\n\t\tfor")
	return s
}
