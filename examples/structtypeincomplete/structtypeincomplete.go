package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func main() {
	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "example.go", src, parser.SkipObjectResolution)

	structType := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.StructType)

	fmt.Printf("fields=%#v incomplete=%#v\n", structType.Fields.List, structType.Incomplete)

	ast.FileExports(f)

	fmt.Printf("fields=%#v incomplete=%#v\n", structType.Fields.List, structType.Incomplete)
}

var src = `package p
type S struct {
	Public  string
	private string
}
`
