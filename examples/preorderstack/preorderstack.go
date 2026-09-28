package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func main() {
	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "example.go", src, parser.Mode(0))

	ast.PreorderStack(f, nil, func(n ast.Node, stack []ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}

		path := make([]string, len(stack))
		for i, outer := range stack {
			path[i] = strings.TrimPrefix(fmt.Sprintf("%T", outer), "*ast.")
		}

		fmt.Printf("%-6s %s\n", ident.Name, strings.Join(path, " > "))
		return true
	})
}

var src = `package p
import _ "log"
func add(n, m int) int {}
`
