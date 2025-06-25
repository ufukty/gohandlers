package list

import (
	"fmt"
	"strings"
)

func quotes(src string) string {
	return fmt.Sprintf("%q", src)
}

func addnewlines(f string) string {
	f = strings.ReplaceAll(f, "}\nfunc", "}\n\nfunc")
	hit := "HandlerInfo"
	f = strings.ReplaceAll(f, fmt.Sprintf("%s{", hit), fmt.Sprintf("%s{\n", hit)) // beginning composite literal
	f = strings.ReplaceAll(f, "}, \"", "},\n\"")                                  // after each line
	f = strings.ReplaceAll(f, "}}", "},\n}")                                      // ending composite literal
	return f
}
