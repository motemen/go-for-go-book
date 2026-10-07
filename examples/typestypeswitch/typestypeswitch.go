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

	for _, name := range pkg.Scope().Names() {
		typ := pkg.Scope().Lookup(name).Type()
		fmt.Printf("%-2s %-18T %-28s %T\n", name, typ, typ, typ.Underlying())
	}
}

var src = `package p

type T struct{ X int }

var b bool
var a [3]int
var s []string
var m map[string]T
var p *T
var c chan<- bool
var f func(int) error
var i interface{ M() }
var t T
`
