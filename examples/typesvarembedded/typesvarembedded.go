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

	st := pkg.Scope().Lookup("S").Type().Underlying().(*types.Struct)

	qf := types.RelativeTo(pkg)
	for field := range st.Fields() {
		fmt.Printf("%s %s Embedded=%v\n",
			field.Name(), types.TypeString(field.Type(), qf), field.Embedded())
	}
}

var src = `package p

type T1 struct{}

type T2 struct{}

type S struct {
	T1
	T2 T2
}
`
