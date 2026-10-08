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

	conf := types.Config{
		Importer: importer.Default(),
		Error: func(err error) {
			fmt.Printf("soft=%-5v %s\n", err.(types.Error).Soft, err)
		},
	}

	_, err := conf.Check("path/to/pkg", fset, []*ast.File{f}, nil)
	fmt.Println(err)
}

var src = `package p

import "log"

func main() {
	var s, t string
	s + 1
	foo = 42
}
`
