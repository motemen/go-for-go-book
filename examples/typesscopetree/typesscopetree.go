package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
)

func main() {
	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "example.go", src, parser.SkipObjectResolution)

	conf := types.Config{Importer: importer.Default()}
	pkg, _ := conf.Check("path/to/pkg", fset, []*ast.File{f}, nil)

	printScope(fset, pkg.Scope(), 0)
}

func printScope(fset *token.FileSet, s *types.Scope, depth int) {
	indent := strings.Repeat("  ", depth)
	// デバッグ情報を取り出す
	name, _, _ := strings.Cut(s.String(), " scope ")
	fmt.Printf("%sscope %s {\n", indent, name)
	for _, name := range s.Names() {
		fmt.Printf("%s  %s\n", indent, s.Lookup(name))
	}
	for child := range s.Children() {
		printScope(fset, child, depth+1)
	}
	fmt.Printf("%s}\n", indent)
}

var src = `package p

import "fmt"

const k = 1

func F[T ~int](x T) {
	var sum int
	for i := range x {
		sum += int(i)
	}
	fmt.Println(sum)
}
`
