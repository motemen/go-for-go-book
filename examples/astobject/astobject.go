package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func main() {
	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "example.go", src, parser.Mode(0))

	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "x" {
			var decl any
			if ident != nil && ident.Obj != nil {
				decl = ident.Obj.Decl
			}
			var kind ast.ObjKind
			if ident.Obj != nil {
				kind = ident.Obj.Kind
			}
			fmt.Printf("%-17sobj=%-12p  kind=%s decl=%T\n", fset.Position(ident.Pos()), ident.Obj, kind, decl)
		}
		return true
	})
}

var src = `package p

import x "pkg"

func f() {
    if x := x.f(); x != nil {
        x(func(x int) int { return x + 1 })
    }
}
`
