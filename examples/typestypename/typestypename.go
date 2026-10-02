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

	objF := pkg.Scope().Lookup("F").(*types.Func)

	objs := []*types.TypeName{
		pkg.Scope().Lookup("T").(*types.TypeName),        // <1>
		pkg.Scope().Lookup("A").(*types.TypeName),        // <2>
		objF.Signature().TypeParams().At(0).Obj(),        // <3>
		types.Universe.Lookup("int").(*types.TypeName),   // <4>
		types.Universe.Lookup("error").(*types.TypeName), // <4>
		types.Universe.Lookup("byte").(*types.TypeName),  // <5>
	}

	for _, obj := range objs {
		fmt.Println(obj)
		fmt.Printf("\tType: %T, IsAlias: %v\n", obj.Type(), obj.IsAlias())
	}
}

var src = `package p

type T struct{}

type A = T

func F[P any](x P) {}
`
