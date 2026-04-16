package post

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"slices"
	"strings"
)

func isThirdParty(s string) bool {
	return strings.Contains(strings.Split(strings.Trim(s, `"`), "/")[0], ".")
}

func imports(f *ast.File) map[token.Pos]actions {
	for _, decl := range f.Decls {
		if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
			for _, sp := range gd.Specs {
				if is, ok := sp.(*ast.ImportSpec); ok && isThirdParty(is.Path.Value) {
					return map[token.Pos]actions{
						is.Pos() - 1: {newline: true},
					}
				}
			}
		}
	}
	return nil
}

// lists the ending positions of type declarations
func typeDecls(f *ast.File) map[token.Pos]actions {
	endings := map[token.Pos]actions{}
	for _, decl := range f.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok {
			endings[fd.End()] = actions{newline: true}
		}
		if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.TYPE {
			endings[gd.End()] = actions{newline: true}
		}
	}
	return endings
}

// lists the beginning/ending positions of entries inside ListHandlers function
func listerEntries(f *ast.File) map[token.Pos]actions {
	endings := map[token.Pos]actions{}
	for _, decl := range f.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok && fd.Name != nil && fd.Name.Name == "ListHandlers" && fd.Body != nil && fd.Body.List != nil {
			for _, stmt := range fd.Body.List {
				if rs, ok := stmt.(*ast.ReturnStmt); ok && rs.Results != nil && len(rs.Results) == 1 {
					if cl, ok := rs.Results[0].(*ast.CompositeLit); ok && cl.Elts != nil && len(cl.Elts) > 0 {
						for _, elt := range cl.Elts {
							endings[elt.Pos()-1] = actions{newline: true}
						}
						endings[cl.Elts[len(cl.Elts)-1].End()-1] = actions{newline: true, comma: true}
					}
				}
			}
		}
	}
	return endings
}

// apply eithers insert a newline only or a comma and newline at each position
func apply(s string, actions map[token.Pos]actions) string {
	bs := []byte(s)
	cs := 0
	for _, pos := range slices.Sorted(maps.Keys(actions)) {
		actions := actions[pos]
		if actions.comma {
			bs = slices.Insert(bs, int(pos)+cs, ',')
			cs++
		}
		if actions.newline {
			bs = slices.Insert(bs, int(pos)+cs, '\n')
			cs++
		}
	}
	return string(bs)
}

func Process(s string) (string, error) {
	f, err := parser.ParseFile(token.NewFileSet(), "", s, parser.AllErrors)
	if err != nil {
		return "", fmt.Errorf("parsing: %w", err)
	}

	s = apply(s, concat(
		imports(f),
		typeDecls(f),
		listerEntries(f),
	))

	return s, nil
}
