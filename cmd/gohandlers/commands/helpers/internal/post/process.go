package post

import (
	"strings"
)

func Process(f string) string {
	f = strings.ReplaceAll(f, "}\nfunc", "}\n\nfunc")
	f = strings.ReplaceAll(f, "HandlerInfo{", "HandlerInfo{\n") // beginning composite literal
	f = strings.ReplaceAll(f, "}, \"", "},\n\"")                // after each line
	f = strings.ReplaceAll(f, "}}", "},\n}")                    // ending composite literal
	return f
}
