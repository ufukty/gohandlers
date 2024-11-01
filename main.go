package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/exp/maps"
)

type Args struct {
	Dir string
	Out string
}

func recvn(s string) string {
	return strings.ToLower(string(s[0:2]))
}

func linearize(n ast.Node) []string {
	literals := []string{}
	ast.Inspect(n, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		switch x := n.(type) {
		case *ast.BasicLit:
			literals = append(literals, x.Value)
		case *ast.Ident:
			literals = append(literals, x.Name)
		default:
			literals = append(literals, fmt.Sprintf("%T", x))
		}
		return true
	})
	return literals
}

var handler = linearize(&ast.FieldList{
	List: []*ast.Field{
		{
			Names: []*ast.Ident{{Name: "w"}},
			Type:  &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "ResponseWriter"}},
		},
		{
			Names: []*ast.Ident{{Name: "r"}},
			Type:  &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "Request"}}},
		},
	},
})

func isHandler(fd *ast.FuncDecl) bool {
	return slices.Compare(linearize(fd.Type.Params), handler) == 0
}

func findHandler(f *ast.File) (*ast.FuncDecl, bool) {
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && isHandler(fd) {
			return fd, true
		}
	}
	return nil, false
}

func findTypeSpec(f *ast.File, n string) (*ast.TypeSpec, bool) {
	for _, d := range f.Decls {
		if gd, ok := d.(*ast.GenDecl); ok && gd.Tok == token.TYPE {
			for _, s := range gd.Specs {
				if ts, ok := s.(*ast.TypeSpec); ok && ts.Name.Name == n {
					return ts, true
				}
			}
		}
	}
	return nil, false
}

func routeparams(ts *ast.TypeSpec) []string {
	ps := []string{}
	if st, ok := ts.Type.(*ast.StructType); ok {
		for _, f := range st.Fields.List {
			if f.Tag != nil {
				st := reflect.StructTag(strings.Trim(f.Tag.Value, "`"))
				if v, ok := st.Lookup("route"); ok {
					ps = append(ps, v)
				}
			}
		}
	}
	return ps
}

var methods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace,
}

func findMethod(fd *ast.FuncDecl) (string, bool) {
	if fd.Doc != nil {
		for _, d := range fd.Doc.List {
			ts := strings.Split(d.Text, " ")
			if len(ts) >= 2 {
				if slices.Index(methods, ts[1]) != -1 {
					return ts[1], true
				}
			}
		}
	}
	return "", false
}

func kebab(input string) string {
	var result strings.Builder
	for i, r := range input {
		if unicode.IsUpper(r) {
			if i != 0 {
				result.WriteRune('-')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

type info struct {
	Method string   `yaml:"method"`
	Path   string   `yaml:"path"`
	Ref    ast.Expr `yaml:"-"`
}

func Main() error {
	args := Args{}

	flag.StringVar(&args.Dir, "dir", "", "the directory contains Go files. one handler and a request binding type is allowed per file")
	flag.StringVar(&args.Out, "out", "register.go", "output file that will be generated in the 'dir'")
	flag.Parse()

	if args.Dir == "" {
		flag.PrintDefaults()
		return fmt.Errorf("missing arguments")
	}

	d, err := parser.ParseDir(token.NewFileSet(), args.Dir, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parsing files in directory: %w", err)
	}

	if len(d) > 1 {
		return fmt.Errorf("found more than one packages: %s", strings.Join(maps.Keys(d), ", "))
	} else if len(d) == 0 {
		return fmt.Errorf("no packages found")
	}
	p := d[maps.Keys(d)[0]]

	infoss := map[string]map[string]info{} // per receiver type
	for _, f := range p.Files {
		if h, ok := findHandler(f); ok {
			bq, ok := findTypeSpec(f, fmt.Sprintf("%sRequest", h.Name.Name))
			if !ok {
				continue
			}

			m, ok := findMethod(h)
			if !ok {
				m = http.MethodGet
			}

			ps := routeparams(bq)
			for i := range ps {
				ps[i] = fmt.Sprintf("{%s}", ps[i])
			}

			path := fmt.Sprintf("/%s", kebab(h.Name.Name))
			if len(ps) > 0 {
				path = fmt.Sprintf("%s/%s", path, strings.Join(ps, "/"))
			}

			var recvt string
			if h.Recv != nil {
				switch t := h.Recv.List[0].Type.(type) {
				case *ast.StarExpr:
					recvt = t.X.(*ast.Ident).Name
				case *ast.Ident:
					recvt = t.Name
				default:
					return fmt.Errorf("unknown type (%T) found in receiver type detection for handler %q", t, h.Name.Name)
				}
			}

			var n ast.Expr
			if h.Recv != nil {
				n = &ast.SelectorExpr{X: &ast.Ident{Name: recvn(recvt)}, Sel: h.Name}
			} else {
				n = h.Name
			}

			i := info{
				Method: m,
				Path:   path,
				Ref:    n,
			}
			if _, ok := infoss[recvt]; !ok {
				infoss[recvt] = map[string]info{}
			}
			infoss[recvt][h.Name.Name] = i
		}
	}

	f := &ast.File{
		Name: maps.Values(p.Files)[0].Name,
		Decls: []ast.Decl{
			&ast.GenDecl{
				Tok:   token.IMPORT,
				Specs: []ast.Spec{&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", "net/http")}}},
			},
			&ast.GenDecl{
				Tok: token.TYPE,
				Specs: []ast.Spec{&ast.TypeSpec{
					Name: &ast.Ident{Name: "HandlerInfo"},
					Type: &ast.StructType{Fields: &ast.FieldList{List: []*ast.Field{
						{
							Names: []*ast.Ident{{Name: "Method"}}, Type: &ast.Ident{Name: "string"},
						},
						{
							Names: []*ast.Ident{{Name: "Path"}}, Type: &ast.Ident{Name: "string"},
						},
						{
							Names: []*ast.Ident{{Name: "Ref"}},
							Type:  &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "HandlerFunc"}},
						},
					}}},
				}},
			},
		},
	}

	fds := []ast.Decl{}
	for recvt, infos := range infoss {
		elts := []ast.Expr{}
		for hn, info := range infos {
			kv := &ast.KeyValueExpr{
				Key: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", hn)},
				Value: &ast.CompositeLit{Elts: []ast.Expr{
					&ast.KeyValueExpr{Key: &ast.Ident{Name: "Method"}, Value: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", info.Method)}},
					&ast.KeyValueExpr{Key: &ast.Ident{Name: "Path"}, Value: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", info.Path)}},
					&ast.KeyValueExpr{Key: &ast.Ident{Name: "Ref"}, Value: info.Ref},
				}},
			}
			elts = append(elts, kv)
		}

		slices.SortFunc(elts, func(a, b ast.Expr) int {
			ka := a.(*ast.KeyValueExpr).Key.(*ast.BasicLit).Value
			kb := b.(*ast.KeyValueExpr).Key.(*ast.BasicLit).Value
			if ka < kb {
				return -1
			} else if ka == kb {
				return 0
			} else {
				return 1
			}
		})

		fd := &ast.FuncDecl{
			Name: &ast.Ident{Name: "ListHandlers"},
			Type: &ast.FuncType{
				Params: &ast.FieldList{List: []*ast.Field{}},
				Results: &ast.FieldList{List: []*ast.Field{
					{Type: &ast.MapType{Key: &ast.Ident{Name: "string"}, Value: &ast.Ident{Name: "HandlerInfo"}}},
				}},
			},
			Body: &ast.BlockStmt{List: []ast.Stmt{
				&ast.ReturnStmt{Results: []ast.Expr{&ast.CompositeLit{
					Type: &ast.MapType{Key: &ast.Ident{Name: "string"}, Value: &ast.Ident{Name: "HandlerInfo"}},
					Elts: elts,
				}}},
			}},
		}

		if recvt != "" {
			fd.Recv = &ast.FieldList{List: []*ast.Field{{
				Names: []*ast.Ident{{Name: recvn(recvt)}},
				Type:  &ast.StarExpr{X: &ast.Ident{Name: recvt}},
			}}}
		}

		fds = append(fds, fd)
	}

	slices.SortFunc(fds, func(a, b ast.Decl) int {
		if a.(*ast.FuncDecl).Recv == nil {
			return -1
		}
		if b.(*ast.FuncDecl).Recv == nil {
			return 1
		}

		at := a.(*ast.FuncDecl).Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
		bt := b.(*ast.FuncDecl).Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name

		if at < bt {
			return -1
		} else if at == bt {
			return 0
		} else {
			return 1
		}
	})
	f.Decls = append(f.Decls, fds...)

	o, err := os.Create(filepath.Join(args.Dir, args.Out))
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer o.Close()
	fmt.Fprint(o, "// Code generated by gohandlers. DO NOT EDIT.\n\n")
	err = format.Node(o, token.NewFileSet(), f)
	if err != nil {
		return fmt.Errorf("printing: %w", err)
	}

	return nil
}

func main() {
	if err := Main(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
