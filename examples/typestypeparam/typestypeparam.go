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

	sig := pkg.Scope().Lookup("Map").(*types.Func).Signature()
	for tp := range sig.TypeParams().TypeParams() {
		fmt.Printf("%s: Index=%d Obj=%v Constraint=%v Underlying=%v\n",
			tp, tp.Index(), tp.Obj(), tp.Constraint(), tp.Underlying())
	}

	x := sig.Params().At(0)
	fmt.Printf("%s: %v (%T)\n", x.Name(), x.Type(), x.Type())
}

var src = `package p

type Number interface{ ~int | ~float64 }

func Map[S any, T Number](xs []S, f func(S) T) []T { return nil }
`
