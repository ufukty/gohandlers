package utilities

import (
	"fmt"
	"go/ast"
	"go/token"
	"gohandlers/pkg/inspects"
)

func quotes(s string) string {
	return fmt.Sprintf("%q", s)
}

func needsJoin(infoss map[inspects.Receiver]map[string]inspects.Info) bool {
	for _, handlers := range infoss {
		for _, info := range handlers {
			if info.RequestType != nil {
				return true
			}
		}
	}
	return false
}

var join = &ast.FuncDecl{
	Name: &ast.Ident{Name: "join"},
	Type: &ast.FuncType{
		Params: &ast.FieldList{List: []*ast.Field{{
			Names: []*ast.Ident{{Name: "segments"}},
			Type:  &ast.Ellipsis{Elt: &ast.Ident{Name: "string"}},
		}}},
		Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "string"}}}},
	},
	Body: &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "url"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: quotes("")}},
			},
			&ast.RangeStmt{
				Key:   &ast.Ident{Name: "i"},
				Value: &ast.Ident{Name: "segment"},
				Tok:   token.DEFINE,
				X:     &ast.Ident{Name: "segments"},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.IfStmt{
							Cond: &ast.BinaryExpr{
								X:  &ast.BinaryExpr{X: &ast.Ident{Name: "i"}, Op: token.NEQ, Y: &ast.BasicLit{Kind: token.INT, Value: "0"}},
								Op: token.LAND,
								Y: &ast.UnaryExpr{
									Op: token.NOT,
									X: &ast.CallExpr{
										Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "strings"}, Sel: &ast.Ident{Name: "HasPrefix"}},
										Args: []ast.Expr{&ast.Ident{Name: "segment"}, &ast.BasicLit{Kind: token.STRING, Value: quotes("/")}},
									},
								},
							},
							Body: &ast.BlockStmt{List: []ast.Stmt{
								&ast.AssignStmt{
									Lhs: []ast.Expr{&ast.Ident{Name: "url"}},
									Tok: token.ADD_ASSIGN,
									Rhs: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: quotes("/")}},
								},
							}},
						},
						&ast.AssignStmt{
							Lhs: []ast.Expr{&ast.Ident{Name: "url"}},
							Tok: token.ADD_ASSIGN,
							Rhs: []ast.Expr{&ast.Ident{Name: "segment"}},
						},
					},
				},
			},
			&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "url"}}},
		},
	},
}

func needsFirstOrZero(infoss map[inspects.Receiver]map[string]inspects.Info) bool {
	for _, handlers := range infoss {
		for _, info := range handlers {
			if info.RequestType != nil && info.RequestType.ContentType == "multipart/form-data" {
				return true
			}
			if info.ResponseType != nil && info.ResponseType.ContentType == "multipart/form-data" {
				return true
			}
		}
	}
	return false
}

var firstOrZero = &ast.FuncDecl{
	Name: &ast.Ident{Name: "firstOrZero"},
	Type: &ast.FuncType{
		TypeParams: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{
						{Name: "E"},
					},
					Type: &ast.Ident{Name: "any"},
				},
			},
		},
		Params: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{
						{Name: "s"},
					},
					Type: &ast.ArrayType{
						Elt: &ast.Ident{Name: "E"},
					},
				},
			},
		},
		Results: &ast.FieldList{
			List: []*ast.Field{
				{
					Type: &ast.Ident{Name: "E"},
				},
			},
		},
	},
	Body: &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.DeclStmt{
				Decl: &ast.GenDecl{
					Tok: token.VAR,
					Specs: []ast.Spec{
						&ast.ValueSpec{
							Names: []*ast.Ident{
								{Name: "v"},
							},
							Type: &ast.Ident{Name: "E"},
						},
					},
				},
			},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X: &ast.CallExpr{
						Fun: &ast.Ident{Name: "len"},
						Args: []ast.Expr{
							&ast.Ident{Name: "s"},
						},
					},
					Op: token.GTR,
					Y:  &ast.BasicLit{Kind: token.INT, Value: "0"},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.AssignStmt{
							Lhs: []ast.Expr{
								&ast.Ident{Name: "v"},
							},
							Tok: token.ASSIGN,
							Rhs: []ast.Expr{
								&ast.IndexExpr{
									X:     &ast.Ident{Name: "s"},
									Index: &ast.BasicLit{Kind: token.INT, Value: "0"},
								},
							},
						},
					},
				},
			},
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.Ident{Name: "v"},
				},
			},
		},
	},
}

func needsFromPart(infoss map[inspects.Receiver]map[string]inspects.Info) bool {
	for _, handlers := range infoss {
		for _, info := range handlers {
			if info.RequestType != nil && len(info.RequestType.Params.Part) > 0 {
				return true
			}
			if info.ResponseType != nil && len(info.ResponseType.Params.Part) > 0 {
				return true
			}
		}
	}
	return false
}

var partReceiver = &ast.GenDecl{
	Tok: token.TYPE,
	Specs: []ast.Spec{
		&ast.TypeSpec{
			Name: &ast.Ident{Name: "partReceiver"},
			Type: &ast.InterfaceType{
				Methods: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								{Name: "FromPart"},
							},
							Type: &ast.FuncType{
								Params: &ast.FieldList{
									List: []*ast.Field{
										{
											Type: &ast.Ident{Name: "string"},
										},
										{
											Type: &ast.Ident{Name: "bool"},
										},
									},
								},
								Results: &ast.FieldList{
									List: []*ast.Field{
										{
											Type: &ast.Ident{Name: "error"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
}

var fromPart = &ast.FuncDecl{
	Name: &ast.Ident{Name: "fromPart"},
	Type: &ast.FuncType{
		Params: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{
						{Name: "dst"},
					},
					Type: &ast.Ident{Name: "partReceiver"},
				},
				{
					Names: []*ast.Ident{
						{Name: "rq"},
					},
					Type: &ast.StarExpr{
						X: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "http"},
							Sel: &ast.Ident{Name: "Request"},
						},
					},
				},
				{
					Names: []*ast.Ident{
						{Name: "key"},
					},
					Type: &ast.Ident{Name: "string"},
				},
			},
		},
		Results: &ast.FieldList{
			List: []*ast.Field{
				{
					Type: &ast.Ident{Name: "error"},
				},
			},
		},
	},
	Body: &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.Ident{Name: "vs"},
					&ast.Ident{Name: "ok"},
				},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.IndexExpr{
						X: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "rq"},
							Sel: &ast.Ident{Name: "PostForm"},
						},
						Index: &ast.Ident{Name: "key"},
					},
				},
			},
			&ast.IfStmt{
				Init: &ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{Name: "err"},
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "dst"},
								Sel: &ast.Ident{Name: "FromPart"},
							},
							Args: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.Ident{Name: "firstOrZero"},
									Args: []ast.Expr{
										&ast.Ident{Name: "vs"},
									},
								},
								&ast.Ident{Name: "ok"},
							},
						},
					},
				},
				Cond: &ast.BinaryExpr{
					X:  &ast.Ident{Name: "err"},
					Op: token.NEQ,
					Y:  &ast.Ident{Name: "nil"},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "fmt"},
										Sel: &ast.Ident{Name: "Errorf"},
									},
									Args: []ast.Expr{
										&ast.BasicLit{Kind: token.STRING, Value: `"FromPart: %w"`},
										&ast.Ident{Name: "err"},
									},
								},
							},
						},
					},
				},
			},
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.Ident{Name: "nil"},
				},
			},
		},
	},
}

func needsToPart(infoss map[inspects.Receiver]map[string]inspects.Info) bool {
	for _, handlers := range infoss {
		for _, info := range handlers {
			if info.RequestType != nil && len(info.RequestType.Params.Part) > 0 {
				return true
			}
			if info.ResponseType != nil && len(info.ResponseType.Params.Part) > 0 {
				return true
			}
		}
	}
	return false
}

var partSupplier = &ast.GenDecl{
	Tok: token.TYPE,
	Specs: []ast.Spec{
		&ast.TypeSpec{
			Name: &ast.Ident{Name: "partSupplier"},
			Type: &ast.InterfaceType{
				Methods: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								{Name: "ToPart"},
							},
							Type: &ast.FuncType{
								Params: &ast.FieldList{},
								Results: &ast.FieldList{
									List: []*ast.Field{
										{
											Type: &ast.Ident{Name: "string"},
										},
										{
											Type: &ast.Ident{Name: "error"},
										},
									},
								},
							},
						},
					},
				},
				Incomplete: false,
			},
		},
	},
}

var toPart = &ast.FuncDecl{
	Name: &ast.Ident{Name: "toPart"},
	Type: &ast.FuncType{
		Params: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{
						{Name: "mp"},
					},
					Type: &ast.StarExpr{
						X: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "multipart"},
							Sel: &ast.Ident{Name: "Writer"},
						},
					},
				},
				{
					Names: []*ast.Ident{
						{Name: "ps"},
					},
					Type: &ast.Ident{Name: "partSupplier"},
				},
				{
					Names: []*ast.Ident{
						{Name: "fieldname"},
					},
					Type: &ast.Ident{Name: "string"},
				},
			},
		},
		Results: &ast.FieldList{
			List: []*ast.Field{
				{
					Type: &ast.Ident{Name: "error"},
				},
			},
		},
	},
	Body: &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.Ident{Name: "v"},
					&ast.Ident{Name: "err"},
				},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "ps"},
							Sel: &ast.Ident{Name: "ToPart"},
						},
					},
				},
			},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X:  &ast.Ident{Name: "err"},
					Op: token.NEQ,
					Y:  &ast.Ident{Name: "nil"},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "fmt"},
										Sel: &ast.Ident{Name: "Errorf"},
									},
									Args: []ast.Expr{
										&ast.BasicLit{Kind: token.STRING, Value: "\"ToPart: %w\""},
										&ast.Ident{Name: "err"},
									},
								},
							},
						},
					},
				},
			},
			&ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.Ident{Name: "w"},
					&ast.Ident{Name: "err"},
				},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "mp"},
							Sel: &ast.Ident{Name: "CreateFormField"},
						},
						Args: []ast.Expr{
							&ast.Ident{Name: "fieldname"},
						},
					},
				},
			},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X:  &ast.Ident{Name: "err"},
					Op: token.NEQ,
					Y:  &ast.Ident{Name: "nil"},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "fmt"},
										Sel: &ast.Ident{Name: "Errorf"},
									},
									Args: []ast.Expr{
										&ast.BasicLit{Kind: token.STRING, Value: "\"CreateFormField: %w\""},
										&ast.Ident{Name: "err"},
									},
								},
							},
						},
					},
				},
			},
			&ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.Ident{Name: "_"},
					&ast.Ident{Name: "err"},
				},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "fmt"},
							Sel: &ast.Ident{Name: "Fprint"},
						},
						Args: []ast.Expr{
							&ast.Ident{Name: "w"},
							&ast.Ident{Name: "v"},
						},
					},
				},
			},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X:  &ast.Ident{Name: "err"},
					Op: token.NEQ,
					Y:  &ast.Ident{Name: "nil"},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "fmt"},
										Sel: &ast.Ident{Name: "Errorf"},
									},
									Args: []ast.Expr{
										&ast.BasicLit{Kind: token.STRING, Value: "\"Fprint: %w\""},
										&ast.Ident{Name: "err"},
									},
								},
							},
						},
					},
				},
			},
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.Ident{Name: "nil"},
				},
			},
		},
	},
}

func needsFromFileHeader(infoss map[inspects.Receiver]map[string]inspects.Info) bool {
	for _, handlers := range infoss {
		for _, info := range handlers {
			if info.RequestType != nil && len(info.RequestType.Params.File) > 0 {
				return true
			}
			if info.ResponseType != nil && len(info.ResponseType.Params.File) > 0 {
				return true
			}
		}
	}
	return false
}

var fileReceiver = &ast.GenDecl{
	Tok: token.TYPE,
	Specs: []ast.Spec{
		&ast.TypeSpec{
			Name: &ast.Ident{Name: "fileReceiver"},
			Type: &ast.InterfaceType{
				Methods: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								{Name: "FromFileHeader"},
							},
							Type: &ast.FuncType{
								Params: &ast.FieldList{
									List: []*ast.Field{
										{
											Type: &ast.StarExpr{
												X: &ast.SelectorExpr{
													X:   &ast.Ident{Name: "multipart"},
													Sel: &ast.Ident{Name: "FileHeader"},
												},
											},
										},
										{
											Type: &ast.Ident{Name: "bool"},
										},
									},
								},
								Results: &ast.FieldList{
									List: []*ast.Field{
										{
											Type: &ast.Ident{Name: "error"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
}

var fromFileHeader = &ast.FuncDecl{
	Name: &ast.Ident{Name: "fromFileHeader"},
	Type: &ast.FuncType{
		Params: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{
						{Name: "dst"},
					},
					Type: &ast.Ident{Name: "fileReceiver"},
				},
				{
					Names: []*ast.Ident{
						{Name: "rq"},
					},
					Type: &ast.StarExpr{
						X: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "http"},
							Sel: &ast.Ident{Name: "Request"},
						},
					},
				},
				{
					Names: []*ast.Ident{
						{Name: "key"},
					},
					Type: &ast.Ident{Name: "string"},
				},
			},
		},
		Results: &ast.FieldList{
			List: []*ast.Field{
				{
					Type: &ast.Ident{Name: "error"},
				},
			},
		},
	},
	Body: &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.Ident{Name: "vs"},
					&ast.Ident{Name: "ok"},
				},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.IndexExpr{
						X: &ast.SelectorExpr{
							X: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "rq"},
								Sel: &ast.Ident{Name: "MultipartForm"},
							},
							Sel: &ast.Ident{Name: "File"},
						},
						Index: &ast.Ident{Name: "key"},
					},
				},
			},
			&ast.IfStmt{
				Init: &ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{Name: "err"},
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "dst"},
								Sel: &ast.Ident{Name: "FromFileHeader"},
							},
							Args: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.Ident{Name: "firstOrZero"},
									Args: []ast.Expr{
										&ast.Ident{Name: "vs"},
									},
								},
								&ast.Ident{Name: "ok"},
							},
						},
					},
				},
				Cond: &ast.BinaryExpr{
					X:  &ast.Ident{Name: "err"},
					Op: token.NEQ,
					Y:  &ast.Ident{Name: "nil"},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "fmt"},
										Sel: &ast.Ident{Name: "Errorf"},
									},
									Args: []ast.Expr{
										&ast.BasicLit{Kind: token.STRING, Value: `"FromFileHeader: %w"`},
										&ast.Ident{Name: "err"},
									},
								},
							},
						},
					},
				},
			},
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.Ident{Name: "nil"},
				},
			},
		},
	},
}

func needsToFile(infoss map[inspects.Receiver]map[string]inspects.Info) bool {
	for _, handlers := range infoss {
		for _, info := range handlers {
			if info.RequestType != nil && len(info.RequestType.Params.File) > 0 {
				return true
			}
			if info.ResponseType != nil && len(info.ResponseType.Params.File) > 0 {
				return true
			}
		}
	}
	return false
}

var quoteEscaper = &ast.GenDecl{
	Tok: token.VAR,
	Specs: []ast.Spec{
		&ast.ValueSpec{
			Names: []*ast.Ident{
				{Name: "quoteEscaper"},
			},
			Values: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "strings"},
						Sel: &ast.Ident{Name: "NewReplacer"},
					},
					Args: []ast.Expr{
						&ast.BasicLit{Kind: token.STRING, Value: "\"\\\\\""},
						&ast.BasicLit{Kind: token.STRING, Value: "\"\\\\\\\\\""},
						&ast.BasicLit{Kind: token.STRING, Value: "`\"`"},
						&ast.BasicLit{Kind: token.STRING, Value: "\"\\\\\\\"\""},
					},
				},
			},
		},
	},
}

var fileSupplier = &ast.GenDecl{
	Tok: token.TYPE,
	Specs: []ast.Spec{
		&ast.TypeSpec{
			Name: &ast.Ident{Name: "fileSupplier"},
			Type: &ast.InterfaceType{
				Methods: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								{Name: "ToFile"},
							},
							Type: &ast.FuncType{
								Params: &ast.FieldList{},
								Results: &ast.FieldList{
									List: []*ast.Field{
										{
											Names: []*ast.Ident{
												{Name: "src"},
											},
											Type: &ast.SelectorExpr{
												X:   &ast.Ident{Name: "io"},
												Sel: &ast.Ident{Name: "Reader"},
											},
										},
										{
											Names: []*ast.Ident{
												{Name: "filename"},
											},
											Type: &ast.Ident{Name: "string"},
										},
										{
											Names: []*ast.Ident{
												{Name: "contentType"},
											},
											Type: &ast.Ident{Name: "string"},
										},
										{
											Names: []*ast.Ident{
												{Name: "err"},
											},
											Type: &ast.Ident{Name: "error"},
										},
									},
								},
							},
						},
					},
				},
				Incomplete: false,
			},
		},
	},
}

var toFile = &ast.FuncDecl{
	Name: &ast.Ident{Name: "toFile"},
	Type: &ast.FuncType{
		Params: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{
						{Name: "mp"},
					},
					Type: &ast.StarExpr{
						X: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "multipart"},
							Sel: &ast.Ident{Name: "Writer"},
						},
					},
				},
				{
					Names: []*ast.Ident{
						{Name: "fs"},
					},
					Type: &ast.Ident{Name: "fileSupplier"},
				},
				{
					Names: []*ast.Ident{
						{Name: "fieldname"},
					},
					Type: &ast.Ident{Name: "string"},
				},
			},
		},
		Results: &ast.FieldList{
			List: []*ast.Field{
				{
					Type: &ast.Ident{Name: "error"},
				},
			},
		},
	},
	Body: &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.Ident{Name: "src"},
					&ast.Ident{Name: "filename"},
					&ast.Ident{Name: "ct"},
					&ast.Ident{Name: "err"},
				},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "fs"},
							Sel: &ast.Ident{Name: "ToFile"},
						},
					},
				},
			},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X:  &ast.Ident{Name: "err"},
					Op: token.NEQ,
					Y:  &ast.Ident{Name: "nil"},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "fmt"},
										Sel: &ast.Ident{Name: "Errorf"},
									},
									Args: []ast.Expr{
										&ast.BasicLit{Kind: token.STRING, Value: "\"ToFile: %w\""},
										&ast.Ident{Name: "err"},
									},
								},
							},
						},
					},
				},
			},
			&ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.Ident{Name: "h"},
				},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.Ident{Name: "make"},
						Args: []ast.Expr{
							&ast.SelectorExpr{
								X:   &ast.Ident{Name: "textproto"},
								Sel: &ast.Ident{Name: "MIMEHeader"},
							},
						},
					},
				},
			},
			&ast.ExprStmt{
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "h"},
						Sel: &ast.Ident{Name: "Set"},
					},
					Args: []ast.Expr{
						&ast.BasicLit{Kind: token.STRING, Value: "\"Content-Disposition\""},
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "fmt"},
								Sel: &ast.Ident{Name: "Sprintf"},
							},
							Args: []ast.Expr{
								&ast.BasicLit{Kind: token.STRING, Value: "`form-data; name=\"%s\"; filename=\"%s\"`"},
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "quoteEscaper"},
										Sel: &ast.Ident{Name: "Replace"},
									},
									Args: []ast.Expr{
										&ast.Ident{Name: "fieldname"},
									},
								},
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "quoteEscaper"},
										Sel: &ast.Ident{Name: "Replace"},
									},
									Args: []ast.Expr{
										&ast.Ident{Name: "filename"},
									},
								},
							},
						},
					},
				},
			},
			&ast.ExprStmt{
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "h"},
						Sel: &ast.Ident{Name: "Set"},
					},
					Args: []ast.Expr{
						&ast.BasicLit{Kind: token.STRING, Value: "\"Content-Type\""},
						&ast.Ident{Name: "ct"},
					},
				},
			},
			&ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.Ident{Name: "dst"},
					&ast.Ident{Name: "err"},
				},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "mp"},
							Sel: &ast.Ident{Name: "CreatePart"},
						},
						Args: []ast.Expr{
							&ast.Ident{Name: "h"},
						},
					},
				},
			},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X:  &ast.Ident{Name: "err"},
					Op: token.NEQ,
					Y:  &ast.Ident{Name: "nil"},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "fmt"},
										Sel: &ast.Ident{Name: "Errorf"},
									},
									Args: []ast.Expr{
										&ast.BasicLit{Kind: token.STRING, Value: "\"CreatePart: %w\""},
										&ast.Ident{Name: "err"},
									},
								},
							},
						},
					},
				},
			},
			&ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.Ident{Name: "_"},
					&ast.Ident{Name: "err"},
				},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "io"},
							Sel: &ast.Ident{Name: "Copy"},
						},
						Args: []ast.Expr{
							&ast.Ident{Name: "dst"},
							&ast.Ident{Name: "src"},
						},
					},
				},
			},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X:  &ast.Ident{Name: "err"},
					Op: token.NEQ,
					Y:  &ast.Ident{Name: "nil"},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "fmt"},
										Sel: &ast.Ident{Name: "Errorf"},
									},
									Args: []ast.Expr{
										&ast.BasicLit{Kind: token.STRING, Value: "\"Copy: %w\""},
										&ast.Ident{Name: "err"},
									},
								},
							},
						},
					},
				},
			},
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.Ident{Name: "nil"},
				},
			},
		},
	},
}

func Produce(infoss map[inspects.Receiver]map[string]inspects.Info) []ast.Decl {
	decls := []ast.Decl{}
	if needsJoin(infoss) {
		decls = append(decls, join)
	}
	if needsFirstOrZero(infoss) {
		decls = append(decls, firstOrZero)
	}
	if needsFromPart(infoss) {
		decls = append(decls, partReceiver, fromPart)
	}
	if needsToPart(infoss) {
		decls = append(decls, partSupplier, toPart)
	}
	if needsFromFileHeader(infoss) {
		decls = append(decls, fileReceiver, fromFileHeader)
	}
	if needsToFile(infoss) {
		decls = append(decls, quoteEscaper, fileSupplier, toFile)
	}
	return decls
}
