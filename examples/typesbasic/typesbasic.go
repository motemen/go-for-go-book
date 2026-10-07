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
		basic := pkg.Scope().Lookup(name).Type().(*types.Basic)
		fmt.Printf("%s: %-14s IsUntyped=%-5v IsNumeric=%-5v Default=%s\n",
			name, basic,
			basic.Info()&types.IsUntyped != 0,
			basic.Info()&types.IsNumeric != 0,
			types.Default(basic))
	}
}

var src = `package p

const a = 1
const b = 'x'
const c = "s"
const d int8 = 1

var x = 1.0
var y = a + 0.5
`
