package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
)

func main() {
	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "example.go", src, parser.Mode(0))

	conf := types.Config{Importer: importer.Default()}

	pkg, _ := conf.Check("path/to/pkg", fset, []*ast.File{f}, nil)

	for _, name := range []string{"A", "B", "S"} {
		alias := pkg.Scope().Lookup(name).Type().(*types.Alias)
		fmt.Printf("%s: Rhs=%v (%T) Unalias=%v (%T)\n",
			alias, alias.Rhs(), alias.Rhs(), types.Unalias(alias), types.Unalias(alias))
	}

	x := pkg.Scope().Lookup("x").Type()
	fmt.Printf("x: %v (%T)\n", x, x)
}

var src = `package p

type T struct{}

type A = T

type B = A

type Set[K comparable] = map[K]bool

type S = Set[string]

var x B
`
