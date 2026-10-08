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

	t := pkg.Scope().Lookup("T").Type().(*types.Named)

	sigs := []*types.Signature{
		pkg.Scope().Lookup("Printf").(*types.Func).Signature(),
		t.Method(0).Signature(),
		pkg.Scope().Lookup("fn").Type().(*types.Signature),
	}

	for _, sig := range sigs {
		fmt.Println(sig)
		fmt.Printf("\tRecv: %v\n", sig.Recv())
		for p := range sig.Params().Variables() {
			fmt.Printf("\tParam: %q %v\n", p.Name(), p.Type())
		}
		for r := range sig.Results().Variables() {
			fmt.Printf("\tResult: %q %v\n", r.Name(), r.Type())
		}
		fmt.Printf("\tVariadic: %v\n", sig.Variadic())
	}
}

var src = `package p

func Printf(format string, args ...any) (n int, err error) { return }

type T struct{}

func (t *T) M(int) bool { return false }

var fn func(string) error
`
