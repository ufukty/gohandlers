package pretty

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"io"
	"strings"

	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/version"
	"github.com/ufukty/gohandlers/cmd/gohandlers/internal/pretty/post"
)

func Print(f *ast.File) (io.Reader, error) {
	b := bytes.NewBuffer([]byte{})
	fmt.Fprint(b, version.Top())
	err := printer.Fprint(b, token.NewFileSet(), f)
	if err != nil {
		return nil, fmt.Errorf("printing: %w", err)
	}
	proccessed := post.Process(b.String())
	formatted, err := format.Source([]byte(proccessed))
	if err != nil {
		return nil, fmt.Errorf("formatting: %w", err)
	}
	return strings.NewReader(string(formatted)), nil
}
