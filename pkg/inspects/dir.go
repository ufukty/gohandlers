package inspects

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/exp/maps"
)

var (
	NOTICE  = "\033[34m" + "notice" + "\033[0m"
	WARNING = "\033[33m" + "warning" + "\033[0m"
	ERROR   = "\033[31m" + "error" + "\033[0m"
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

func findHandlers(f *ast.File) []*ast.FuncDecl {
	hs := []*ast.FuncDecl{}
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && isHandler(fd) {
			hs = append(hs, fd)
		}
	}
	return hs
}

type BindingTypeParameterSources struct {
	Route, Query map[string]string // Header
	Json, Form   map[string]string // Body
}

type BindingTypeInfo struct {
	Typename     string
	ContainsBody bool
	Empty        bool
	ContentType  string
	Params       BindingTypeParameterSources // param -> fieldname
}

func bti(rqtn string, ts *ast.TypeSpec) (*BindingTypeInfo, error) {
	bti := &BindingTypeInfo{
		Typename: rqtn,
		Params: BindingTypeParameterSources{
			Route: map[string]string{},
			Query: map[string]string{},
			Json:  map[string]string{},
			Form:  map[string]string{},
		},
	}

	if st, ok := ts.Type.(*ast.StructType); ok {
		for _, f := range st.Fields.List {
			if f.Tag != nil {
				st := reflect.StructTag(strings.Trim(f.Tag.Value, "`"))
				if v, ok := st.Lookup("route"); ok {
					bti.Params.Route[v] = f.Names[0].Name
				}
				if v, ok := st.Lookup("query"); ok {
					bti.Params.Query[v] = f.Names[0].Name
				}
				if v, ok := st.Lookup("json"); ok {
					bti.Params.Json[v] = f.Names[0].Name
				}
				if v, ok := st.Lookup("form"); ok {
					bti.Params.Form[v] = f.Names[0].Name
				}
			}
		}
	}

	containsHeaderParams := len(bti.Params.Route) > 0 || len(bti.Params.Query) > 0
	bti.ContainsBody = len(bti.Params.Json) > 0 || len(bti.Params.Form) > 0
	bti.Empty = !bti.ContainsBody && !containsHeaderParams

	if len(bti.Params.Json) > 0 && len(bti.Params.Form) > 0 {
		return nil, fmt.Errorf("determining Content Type for body: both json {%s} and form {%s} tagged fields found",
			strings.Join(maps.Values(bti.Params.Json), ", "), strings.Join(maps.Values(bti.Params.Form), ", "),
		)
	}

	switch {
	case len(bti.Params.Json) > 0:
		bti.ContentType = "application/json"
	case len(bti.Params.Form) > 0:
		bti.ContentType = "application/x-www-form-urlencoded"
	}

	return bti, nil
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
	return strings.ToLower(string(s[0:min(2, len(s))]))
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

func checkBodyForIdent(h *ast.FuncDecl, i *ast.Ident) bool {
	found := false
	ast.Inspect(h.Body, func(n ast.Node) bool {
		if i2, ok := n.(*ast.Ident); ok && i.Name == i2.Name {
			found = true
		}
		return !found
	})
	return found
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

type Mode string

func (m Mode) Ignore() bool {
	return m == "ignore"
}

func (m Mode) ToList() bool {
	return m != "ignore" && m == "list"
}

func (m Mode) ParseBindings() bool {
	return m != "ignore" && m != "list"
}

type Doc struct {
	Method, Path string
	Mode
}

var whitespaces = regexp.MustCompile(`\s+`)

func parseDoc(fd *ast.FuncDecl) Doc {
	doc := Doc{}
	if fd.Doc != nil {
		for _, c := range fd.Doc.List {
			line := c.Text
			line = strings.TrimPrefix(line, "//")
			line = strings.TrimPrefix(line, "/*")
			line = strings.TrimPrefix(line, "*")
			line = strings.TrimSpace(line)
			line = whitespaces.ReplaceAllString(line, " ")
			for i, word := range strings.Split(line, " ") {
				switch {
				case strings.HasPrefix(word, "gh:") && i == 0:
					doc.Mode = Mode(strings.TrimPrefix(word, "gh:"))
				case slices.Contains(methods, word) && i == 0:
					doc.Method = word
				case strings.HasPrefix(word, "/") && i <= 1:
					doc.Path = word
				}
			}
		}
	}
	return doc
}

func has[K comparable, V any](m map[K]V, k K) bool {
	_, ok := m[k]
	return ok
}

var titles = regexp.MustCompile(`([A-Z][a-z]+)\S*`)

var methodMap = map[string]string{
	"Get":   "GET",
	"Visit": "GET", //

	"Head": "HEAD",

	"Post":   "POST",
	"Upload": "POST", //
	"Create": "POST", //

	"Put":     "PUT",
	"Replace": "PUT", //

	"Patch":  "PATCH",
	"Update": "PATCH", //

	"Delete": "DELETE",
	"Remove": "DELETE", //

	"Connect": "CONNECT",
	"Options": "OPTIONS",
	"Trace":   "TRACE",
}

func decideMethodFromHandlerName(h *ast.FuncDecl) string {
	matches := titles.FindStringSubmatch(h.Name.Name)
	if len(matches) > 1 && has(methodMap, matches[1]) {
		return methodMap[matches[1]]
	}
	return ""
}

func decideMethodFromRequest(rti *BindingTypeInfo) string {
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

// ordered by precedence
func electMethod(docComment, handlerName, requestBinding string) string {
	return cmp.Or(docComment, handlerName, requestBinding, string(http.MethodGet))
}

func handlerMethod(h *ast.FuncDecl, doc Doc, rti *BindingTypeInfo, filename string) (string, []string) {
	fromBindingType := ""
	if rti != nil && doc.ParseBindings() {
		fromBindingType = decideMethodFromRequest(rti)
	}
	fromHandlerName := decideMethodFromHandlerName(h)
	method := electMethod(doc.Method, fromHandlerName, fromBindingType)

	okDoc := doc.Method != ""
	okName := fromHandlerName != ""
	okBq := fromBindingType != ""

	complaints := []string{}

	if okDoc && okName {
		if doc.Method != fromHandlerName {
			complaints = append(complaints, fmt.Sprintf("%s: %s:%s: handler name implies %q but doc comment specifies %q", WARNING, filename, h.Name.Name, fromHandlerName, doc.Method))
		}
	}

	if doc.ParseBindings() {
		// implicit assignment notices
		// TODO: decide if assigning method by handler name prefix implicit?
		if !okDoc && !okName {
			if okBq {
				complaints = append(complaints, fmt.Sprintf("%s: %s:%s: implicitly assigned %q based on if the request contains a body", NOTICE, filename, h.Name.Name, method))
			} else {
				complaints = append(complaints, fmt.Sprintf("%s: %s:%s: implicitly assigned %q without any information", NOTICE, filename, h.Name.Name, method))
			}
		}

		if !slices.Contains(bodied, method) && slices.Contains(bodied, fromBindingType) {
			complaints = append(complaints, fmt.Sprintf("%s: %s:%s: assigned %q but the request binding type contains a body", ERROR, filename, h.Name.Name, method))
		}
		if slices.Contains(bodied, method) && !slices.Contains(bodied, fromBindingType) {
			complaints = append(complaints, fmt.Sprintf("%s: %s:%s: assigned %q but the request binding type doesn't contain a body", ERROR, filename, h.Name.Name, method))
		}
	}

	return method, complaints
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

func handlerPathFromBindingType(h *ast.FuncDecl, rti *BindingTypeInfo) string {
	ps := []string{}
	if rti != nil {
		for i := range rti.Params.Route {
			ps = append(ps, fmt.Sprintf("{%s}", i))
		}
	}

	path := fmt.Sprintf("/%s", kebab(h.Name.Name))
	if len(ps) > 0 {
		slices.Sort(ps)
		path = fmt.Sprintf("%s/%s", path, strings.Join(ps, "/"))
	}
	return path
}

func checkHandlerPathInDoc(doc Doc, rti *BindingTypeInfo) (missing []string) {
	if rti == nil || rti.Params.Route == nil {
		return
	}
	words := strings.Split(strings.TrimPrefix(doc.Path, "/"), "/")
	for _, param := range maps.Keys(rti.Params.Route) {
		if !slices.Contains(words, fmt.Sprintf("{%s}", param)) {
			missing = append(missing, param)
		}
	}
	return
}

func handlerPath(h *ast.FuncDecl, doc Doc, rti *BindingTypeInfo, filename string) (string, string) {
	if doc.Path == "" {
		return handlerPathFromBindingType(h, rti), ""
	}
	missings := checkHandlerPathInDoc(doc, rti)
	if len(missings) > 0 {
		complaint := fmt.Sprintf("%s: %s:%s: the path specified in doc comment has been added missing route parameters: %s", NOTICE, filename, h.Name.Name, strings.Join(missings, ", "))
		suffix := ""
		for _, missing := range missings {
			suffix += fmt.Sprintf("/{%s}", missing)
		}
		return filepath.Join(doc.Path, suffix), complaint
	}
	return doc.Path, ""
}

type Receiver struct {
	Name, Type string
}

type Info struct {
	Method       string
	Path         string
	Ref          ast.Expr
	RequestType  *BindingTypeInfo
	ResponseType *BindingTypeInfo
}

func Dir(dir string, verbose bool) (map[Receiver]map[string]Info, string, error) {
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
	for fn, f := range p.Files {
		for _, h := range findHandlers(f) {
			doc := parseDoc(h)
			if doc.Mode.Ignore() {
				if verbose {
					fmt.Fprintf(os.Stderr, "%s: ignoring %s:%s\n", NOTICE, fn, h.Name.Name)
				}
				continue
			}

			recvt, err := receiverType(h)
			if err != nil {
				return nil, "", fmt.Errorf("inspecting receiver type of handler: %w", err)
			}
			i := Info{
				Ref: ref(h, recvt),
			}

			if doc.Mode.ParseBindings() {
				bqtn := fmt.Sprintf("%sRequest", h.Name.Name)
				bq, ok := findTypeSpec(f, bqtn)
				if ok && checkBodyForIdent(h, bq.Name) {
					i.RequestType, err = bti(bqtn, bq)
					if err != nil {
						return nil, "", fmt.Errorf("inspecting request binding type: %w", err)
					}
				}
			}

			method, complaints := handlerMethod(h, doc, i.RequestType, fn)
			for _, complaint := range complaints {
				if verbose || strings.HasPrefix(complaint, ERROR) || strings.HasPrefix(complaint, WARNING) {
					fmt.Fprintln(os.Stderr, complaint)
				}
			}
			i.Method = method

			path, complaint := handlerPath(h, doc, i.RequestType, fn)
			if complaint != "" && verbose {
				fmt.Fprintln(os.Stderr, complaint)
			}
			i.Path = path

			if doc.Mode.ParseBindings() {
				bstn := fmt.Sprintf("%sResponse", h.Name.Name)
				bs, ok := findTypeSpec(f, bstn)
				if ok && checkBodyForIdent(h, bs.Name) {
					i.ResponseType, err = bti(bstn, bs)
					if err != nil {
						return nil, "", fmt.Errorf("inspecting response binding type: %w", err)
					}
				}
			}

			if verbose {
				fmt.Printf("adding %s %s for %s\n", i.Method, i.Path, h.Name.Name)
			}
			r := Receiver{recvn(recvt), recvt}
			if _, ok := infoss[r]; !ok {
				infoss[r] = map[string]Info{}
			}
			infoss[r][h.Name.Name] = i
		}
	}

	return infoss, maps.Values(p.Files)[0].Name.Name, nil
}
