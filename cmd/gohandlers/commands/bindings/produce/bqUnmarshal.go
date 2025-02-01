package produce

import (
	"go/ast"
	"gohandlers/pkg/inspects"
)

func BqUmarshal(i inspects.Info) []ast.Decl {
	switch i.RequestType.ContentType {
	case "application/x-www-form-urlencode":
		panic("to implement")
	case "multipart/form-data":
		return []ast.Decl{ResponseUnmarshalMultipartFormData(i)}
	default:
		return []ast.Decl{}
	}
}
