package post

import (
	"regexp"
	"strings"
)

var imports = regexp.MustCompile(`(?m)import \(((?:\s+"[\w]+(?:/[\w.]+)*"\n)*)((?:\s+"[\w.]+(?:/[\w.]+)*"\n)*)\)`)

func Process(f string) string {
	f = strings.ReplaceAll(f, "}\nfunc", "}\n\nfunc")
	f = strings.ReplaceAll(f, "HandlerInfo{", "HandlerInfo{\n") // beginning composite literal
	f = strings.ReplaceAll(f, "}, \"", "},\n\"")                // after each line
	f = strings.ReplaceAll(f, "}}", "},\n}")                    // ending composite literal
	f = imports.ReplaceAllString(f, "import ($1\n$2)")          // split packages starts with domains
	return f
}
