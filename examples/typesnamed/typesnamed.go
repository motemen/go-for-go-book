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

	for _, name := range []string{"Celsius", "Point", "Celsius2"} {
		named := pkg.Scope().Lookup(name).Type().(*types.Named)
		fmt.Printf("%s: Obj=%v Underlying=%v\n", named, named.Obj(), named.Underlying())
		for m := range named.Methods() {
			fmt.Printf("\tmethod %s\n", m.Name())
		}
	}

	// ジェネリックな型のインスタンス
	v := pkg.Scope().Lookup("v").Type().(*types.Named)
	fmt.Printf("%s: Origin=%v TypeParams=%v TypeArgs=%v\n",
		v, v.Origin(), v.Origin().TypeParams().At(0), v.TypeArgs().At(0))
}

var src = `package p

type Celsius float64

func (c Celsius) String() string { return "" }

type Point struct{ X, Y int }

func (p *Point) Move(dx, dy int) {}

type Celsius2 Celsius

type List[T any] []T

var v List[string]
`
