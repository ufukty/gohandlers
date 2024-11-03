package inspects

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"reflect"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/exp/maps"
)

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

type RequestTypeInfo struct {
	Typename     string
	RouteParams  map[string]string // route-param -> field-name
	ContainsBody bool
}

func routeparams(ts *ast.TypeSpec) map[string]string {
	ps := map[string]string{}
	if st, ok := ts.Type.(*ast.StructType); ok {
		for _, f := range st.Fields.List {
			if f.Tag != nil {
				st := reflect.StructTag(strings.Trim(f.Tag.Value, "`"))
				if v, ok := st.Lookup("route"); ok {
					ps[v] = f.Names[0].Name
				}
			}
		}
	}
	return ps
}

func containsbody(ts *ast.TypeSpec) bool {
	if st, ok := ts.Type.(*ast.StructType); ok {
		for _, f := range st.Fields.List {
			if f.Tag != nil {
				st := reflect.StructTag(strings.Trim(f.Tag.Value, "`"))
				if _, ok := st.Lookup("json"); ok {
					return true
				}
			}
		}
	}
	return false
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

type Info struct {
	Method      string
	Path        string
	Ref         ast.Expr
	RequestType *RequestTypeInfo
}

type Receiver struct {
	Name, Type string
}

func Dir(dir string) (map[Receiver]map[string]Info, string, error) {
	d, err := parser.ParseDir(token.NewFileSet(), dir, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, "", fmt.Errorf("parsing files in directory: %w", err)
	}

	if len(d) > 1 {
		return nil, "", fmt.Errorf("found more than one packages: %s", strings.Join(maps.Keys(d), ", "))
	} else if len(d) == 0 {
		return nil, "", fmt.Errorf("no packages found")
	}
	p := d[maps.Keys(d)[0]]

	infoss := map[Receiver]map[string]Info{}
	for _, f := range p.Files {
		if h, ok := findHandler(f); ok {
			rbtn := fmt.Sprintf("%sRequest", h.Name.Name)
			bq, ok := findTypeSpec(f, rbtn)
			if !ok {
				continue
			}

			rti := &RequestTypeInfo{
				Typename:     rbtn,
				RouteParams:  map[string]string{},
				ContainsBody: containsbody(bq),
			}

			m, ok := findMethod(h)
			if !ok {
				if rti.ContainsBody {
					m = http.MethodPost
				} else {
					m = http.MethodGet
				}
				fmt.Fprintf(os.Stderr, "notice: method %q implicitly assigned to %q\n", m, h.Name.Name)
			}
			fmt.Printf("adding %s %s...\n", m, h.Name.Name)

			rti.RouteParams = routeparams(bq)
			ps := []string{}
			for i := range rti.RouteParams {
				ps = append(ps, fmt.Sprintf("{%s}", i))
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
					return nil, "", fmt.Errorf("unknown type (%T) found in receiver type detection for handler %q", t, h.Name.Name)
				}
			}

			var n ast.Expr
			if h.Recv != nil {
				n = &ast.SelectorExpr{X: &ast.Ident{Name: recvn(recvt)}, Sel: h.Name}
			} else {
				n = h.Name
			}

			r := Receiver{recvn(recvt), recvt}
			i := Info{
				Method:      m,
				Path:        path,
				Ref:         n,
				RequestType: rti,
			}
			if _, ok := infoss[r]; !ok {
				infoss[r] = map[string]Info{}
			}
			infoss[r][h.Name.Name] = i
		}
	}

	return infoss, maps.Values(p.Files)[0].Name.Name, nil
}
