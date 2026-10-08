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

	for n := range ast.Preorder(f) {
		if ident, ok := n.(*ast.Ident); ok {
			fmt.Println(ident.Name)
		}
	}
}

var src = `package p
import _ "log"
func add(n, m int) {}
`
