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
	f, _ := parser.ParseFile(fset, "example.go", src, parser.SkipObjectResolution)

	conf := types.Config{Importer: importer.Default()}

	pkg, _ := conf.Check("path/to/pkg", fset, []*ast.File{f}, nil)
	fmt.Println(pkg.Scope().Lookup("c2").(*types.Const).Val())
}

var src = `package p

const (
	s = "Hello, " + "world"
	c1 = complex(iota, float64(len(s)))
	c2
)
`
