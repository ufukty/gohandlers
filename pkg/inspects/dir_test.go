package inspects

import (
	"fmt"
	"go/ast"
	"maps"
	"net/http"
	"os"
	"slices"
	"strings"
	"testing"
)

const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	PATCH  = "PATCH"
	DELETE = "DELETE"
)

func TestParseDoc(t *testing.T) {
	type tc struct {
		description string
		input       []string
		output      Doc
	}
	tcs := []tc{
		{
			description: "ignore",
			input:       []string{"// gh:ignore"},
			output:      Doc{Mode: "ignore"},
		},
		{
			description: "ignore with whitespaces",
			input:       []string{"//    gh:ignore"},
			output:      Doc{Mode: "ignore"},
		},
		{
			description: "only the method",
			input:       []string{"// GET"},
			output:      Doc{Method: GET},
		},
		{
			description: "only the method with whitespaces",
			input:       []string{"//   GET   "},
			output:      Doc{Method: GET},
		},
		{
			description: "only the path",
			input:       []string{"// /index.html"},
			output:      Doc{Path: "/index.html"},
		},
		{
			description: "only the path with whitespaces",
			input:       []string{"//     /index.html   "},
			output:      Doc{Path: "/index.html"},
		},
		{
			description: "method and path",
			input:       []string{"// GET /index.html"},
			output:      Doc{Method: GET, Path: "/index.html"},
		},
		{
			description: "method and path with whitespaces",
			input:       []string{"//   GET   /index.html    "},
			output:      Doc{Method: GET, Path: "/index.html"},
		},
		{
			description: "ignore, method and path",
			input:       []string{"// gh:ignore", "// GET /index.html"},
			output:      Doc{Method: GET, Path: "/index.html", Mode: "ignore"},
		},
		{
			description: "ignore, method and path with whitespaces",
			input:       []string{"//   gh:ignore    ", "// GET     /index.html   "},
			output:      Doc{Method: GET, Path: "/index.html", Mode: "ignore"},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			list := []*ast.Comment{}
			for _, line := range tc.input {
				list = append(list, &ast.Comment{Text: line})
			}
			fd := &ast.FuncDecl{Doc: &ast.CommentGroup{List: list}}
			got := parseDoc(fd)
			if got.Mode.Ignore() != tc.output.Mode.Ignore() {
				t.Errorf(".Ignore: expected '%v' got '%v'", tc.output.Mode.Ignore(), got.Mode.Ignore())
			}
			if got.Method != tc.output.Method {
				t.Errorf(".Method: expected '%v' got '%v'", tc.output.Method, got.Method)
			}
			if got.Path != tc.output.Path {
				t.Errorf(".Path: expected '%v' got '%v'", tc.output.Path, got.Path)
			}
		})
	}
}

func TestDecideMethodFromHandlerName(t *testing.T) {
	type input string
	type output string
	type tc struct {
		output
		input
	}
	handlers := []tc{
		{output(""), input("User")},
		{output(http.MethodDelete), input("DeleteUser")},
		{output(http.MethodDelete), input("RemoveUser")},
		{output(http.MethodGet), input("Visit")},
		{output(http.MethodPatch), input("PatchUser")},
		{output(http.MethodPatch), input("UpdateUser")},
		{output(http.MethodPost), input("Create")},
		{output(http.MethodPost), input("CreateUser")},
		{output(http.MethodPost), input("PostUser")},
		{output(http.MethodPut), input("PutUser")},
		{output(http.MethodPut), input("ReplaceUser")},
	}
	for _, tc := range handlers {
		t.Run(string(tc.input), func(t *testing.T) {
			method := decideMethodFromHandlerName(&ast.FuncDecl{Name: ast.NewIdent(string(tc.input))})
			if string(tc.output) != method {
				t.Fatalf("expected %q got %q", tc.output, method)
			}
		})
	}
}

func TestHandlerMethodComplaints(t *testing.T) {
	type RequestBindingTypeStatus string
	const (
		Absent      RequestBindingTypeStatus = "absent"
		WithBody    RequestBindingTypeStatus = "with-body"
		WithoutBody RequestBindingTypeStatus = "without-body"
	)
	type input struct {
		docComment     string
		handlerName    string
		requestBinding RequestBindingTypeStatus
	}
	type contains []string
	type tc struct {
		description string
		input
		contains
	}
	tcs := []tc{
		{"", input{"", "User", Absent}, contains{"without any information"}},
		{"", input{"", "User", WithBody}, contains{"based on if the request contains a body"}},
		{"", input{"", "User", WithoutBody}, contains{"based on if the request contains a body"}},
		{"", input{"", "Visit", WithBody}, contains{"but the request binding type contains a body"}},
		{"", input{"", "Create", WithoutBody}, contains{"but the request binding type doesn't contain a body"}},
		{"", input{"", "Create", Absent}, contains{"but the request binding type doesn't contain a body"}},
		{"", input{GET, "Visit", WithBody}, contains{"but the request binding type contains a body"}},
		{"", input{POST, "User", WithoutBody}, contains{"but the request binding type doesn't contain a body"}},
		{"", input{POST, "Visit", Absent}, contains{"but the request binding type doesn't contain a body", "name implies"}},
		{"", input{POST, "Visit", WithBody}, contains{"name implies"}},
	}

	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			var bti *BindingTypeInfo
			switch tc.requestBinding {
			case WithoutBody:
				bti = &BindingTypeInfo{ContainsBody: false}
			case WithBody:
				bti = &BindingTypeInfo{ContainsBody: true}
			}
			h := &ast.FuncDecl{Name: ast.NewIdent(tc.handlerName)}
			if tc.docComment != "" {
				h.Doc = &ast.CommentGroup{List: []*ast.Comment{{Text: fmt.Sprintf("// %s", tc.docComment)}}}
			}
			doc := parseDoc(h)
			_, complaints := handlerMethod(h, doc, bti, "")
			for _, expectation := range tc.contains {
				if !slices.ContainsFunc(complaints, func(complaint string) bool { return strings.Contains(complaint, expectation) }) {
					t.Errorf("method: expected to contain %q, got %q", expectation, complaints)
				}
			}
		})
	}
}

func TestHandlerMethod(t *testing.T) {
	type RequestBindingTypeStatus string
	const (
		Absent      RequestBindingTypeStatus = "absent"
		WithBody    RequestBindingTypeStatus = "with-body"
		WithoutBody RequestBindingTypeStatus = "without-body"
	)
	type input struct {
		docComment     string
		handlerName    string
		requestBinding RequestBindingTypeStatus
	}
	type output struct {
		expected string
		complain bool
	}
	type tc struct {
		description string
		input
		output
	}

	tcs := []tc{
		{"binding with body", input{"", "User", WithBody}, output{POST, true}},
		{"binding without body", input{"", "User", WithoutBody}, output{GET, true}},
		{"doc conflicting with binding", input{POST, "User", WithoutBody}, output{POST, true}},
		{"doc conflicting with handler", input{POST, "Visit", Absent}, output{POST, true}},
		{"doc conflicting with handler", input{POST, "Visit", WithoutBody}, output{POST, true}},
		{"doc+binding conflicting with handler", input{POST, "Visit", WithBody}, output{POST, true}},
		{"doc+handler conflicting with binding", input{GET, "Visit", WithBody}, output{GET, true}},
		{"handler conflicting with binding", input{"", "Visit", WithBody}, output{GET, true}},
		{"not enough information", input{"", "User", Absent}, output{GET, true}},
	}

	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			var bti *BindingTypeInfo
			switch tc.requestBinding {
			case WithoutBody:
				bti = &BindingTypeInfo{ContainsBody: false}
			case WithBody:
				bti = &BindingTypeInfo{ContainsBody: true}
			}
			h := &ast.FuncDecl{Name: ast.NewIdent(tc.handlerName)}
			if tc.docComment != "" {
				h.Doc = &ast.CommentGroup{List: []*ast.Comment{{Text: fmt.Sprintf("// %s", tc.docComment)}}}
			}
			doc := parseDoc(h)
			method, complaints := handlerMethod(h, doc, bti, "")
			if method != tc.expected {
				t.Errorf("method: expected %q, got %q", tc.expected, method)
			}
			if (complaints != nil) != tc.complain {
				t.Errorf("complaints: expected %v, got %v", tc.complain, complaints != nil)
				t.Log(strings.Join(complaints, "\n"))
			}
		})
	}
}

func TestDir_petstore(t *testing.T) {
	os.Stderr, _ = os.Open(os.DevNull) // silence printed errors

	infoss, _, err := Dir("testdata/petstore", false)
	if err != nil {
		t.Fatalf("act: Dir: %v", err)
	}

	petstore := first(maps.Values(infoss))
	if l := len(petstore); l != 4 {
		t.Fatalf("expected 4 got %d", l)
	}
}
