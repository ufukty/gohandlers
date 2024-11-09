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

type RequestTypeInfo struct {
	Typename     string
	RouteParams  map[string]string // route-param -> field-name
	QueryParams  map[string]string // query-param -> field-name
	ContainsBody bool
}

func rti(rqtn string, ts *ast.TypeSpec) *RequestTypeInfo {
	rti := &RequestTypeInfo{
		Typename:     rqtn,
		RouteParams:  map[string]string{},
		QueryParams:  map[string]string{},
		ContainsBody: false,
	}

	if st, ok := ts.Type.(*ast.StructType); ok {
		for _, f := range st.Fields.List {
			if f.Tag != nil {
				st := reflect.StructTag(strings.Trim(f.Tag.Value, "`"))
				if v, ok := st.Lookup("route"); ok {
					rti.RouteParams[v] = f.Names[0].Name
				}
				if v, ok := st.Lookup("query"); ok {
					rti.QueryParams[v] = f.Names[0].Name
				}
				if _, ok := st.Lookup("json"); ok {
					rti.ContainsBody = true
				}
			}
		}
	}

	return rti
}

func receiverType(h *ast.FuncDecl) (string, error) {
	if h.Recv == nil {
		return "", nil
	}
	switch t := h.Recv.List[0].Type.(type) {
	case *ast.StarExpr:
		return t.X.(*ast.Ident).Name, nil
	case *ast.Ident:
		return t.Name, nil
	default:
		return "", fmt.Errorf("unknown type (%T) found in receiver type detection for handler %q", t, h.Name.Name)
	}
}

func recvn(s string) string {
	return strings.ToLower(string(s[0:2]))
}

func ref(h *ast.FuncDecl, recvt string) ast.Expr {
	if h.Recv != nil {
		return &ast.SelectorExpr{X: &ast.Ident{Name: recvn(recvt)}, Sel: h.Name}
	}
	return h.Name
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

func findMethodInDocs(fd *ast.FuncDecl) (string, bool) {
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

func decideMethodFromRequest(rti *RequestTypeInfo) string {
	if rti.ContainsBody {
		return http.MethodPost
	}
	return http.MethodGet
}

var bodied = []string{
	http.MethodPatch,
	http.MethodPost,
	http.MethodPut,
}

func handlerMethod(h *ast.FuncDecl, rti *RequestTypeInfo) string {
	mDoc, okDoc := findMethodInDocs(h)

	var mBq string
	var okBq bool
	if rti != nil {
		mBq = decideMethodFromRequest(rti)
		okBq = true
	}

	switch {
	case okDoc && okBq:
		if mDoc == http.MethodGet && mBq != http.MethodGet {
			fmt.Fprintf(os.Stderr, "warning: handler %q explicitly assigned %q but the request contains a body\n", h.Name.Name, http.MethodGet)
		}
		if slices.Contains(bodied, mDoc) && mBq == http.MethodGet {
			fmt.Fprintf(os.Stderr, "warning: handler %q explicitly assigned %q but the request doesn't contains a body\n", h.Name.Name, mDoc)
		}
		return mDoc

	case okDoc && !okBq:
		return mDoc

	case !okDoc && okBq:
		fmt.Fprintf(os.Stderr, "notice: handler %q implicitly assigned %q because of the request contains body\n", h.Name.Name, mBq)
		return mBq

	case !okDoc && !okBq:
		fmt.Fprintf(os.Stderr, "warning: handler %q is assigned %q even though there is not enough information to decide\n", h.Name.Name, http.MethodGet)
		return http.MethodGet
	}

	return "" // can't reach here
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

func handlerPath(h *ast.FuncDecl, rti *RequestTypeInfo) string {
	ps := []string{}
	if rti != nil {
		for i := range rti.RouteParams {
			ps = append(ps, fmt.Sprintf("{%s}", i))
		}
	}

	path := fmt.Sprintf("/%s", kebab(h.Name.Name))
	if len(ps) > 0 {
		path = fmt.Sprintf("%s/%s", path, strings.Join(ps, "/"))
	}
	return path

}

type Receiver struct {
	Name, Type string
}

type Info struct {
	Method      string
	Path        string
	Ref         ast.Expr
	RequestType *RequestTypeInfo
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
			recvt, err := receiverType(h)
			if err != nil {
				return nil, "", fmt.Errorf("inspecting receiver type of handler: %w", err)
			}
			i := Info{
				Ref: ref(h, recvt),
			}

			bqtn := fmt.Sprintf("%sRequest", h.Name.Name)
			bq, ok := findTypeSpec(f, bqtn)
			if ok {
				i.RequestType = rti(bqtn, bq)
			}

			i.Method = handlerMethod(h, i.RequestType)
			i.Path = handlerPath(h, i.RequestType)

			fmt.Printf("adding %s %s for %s\n", i.Method, i.Path, h.Name.Name)
			r := Receiver{recvn(recvt), recvt}
			if _, ok := infoss[r]; !ok {
				infoss[r] = map[string]Info{}
			}
			infoss[r][h.Name.Name] = i
		}
	}

	return infoss, maps.Values(p.Files)[0].Name.Name, nil
}
