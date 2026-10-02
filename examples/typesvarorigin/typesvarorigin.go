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

	// わかりやすさのため type assertion
	gObj := pkg.Scope().Lookup("g").(*types.Var)
	fmt.Println(gObj)

	// Gen[string]
	gType := gObj.Type().(*types.Named)

	// Gen[string] のフィールド
	for v := range gType.Underlying().(*types.Struct).Fields() {
		printVar(fset, v)
	}

	// Gen[string] のメソッド Meth の引数
	for v := range gType.Method(0).Signature().Params().Variables() {
		printVar(fset, v)
	}
}

func printVar(fset *token.FileSet, v *types.Var) {
	fmt.Println(v)
	if v.Origin() != v {
		fmt.Printf("\tOrigin: %s (%s)\n", v.Origin(), fset.Position(v.Origin().Pos()))
	}
}

var src = `package p

type Gen[T any] struct {
	FieldT T
	FieldNum int
}

func (Gen[T]) Meth(argT T, argNum int) {}

var g Gen[string]
`
