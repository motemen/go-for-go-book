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

	info := types.Info{
		Scopes: map[ast.Node]*types.Scope{},
	}
	_, _ = conf.Check("path/to/pkg", fset, []*ast.File{f}, &info)

	objPkgName := info.Scopes[f].Lookup("fmtPkg").(*types.PkgName)

	fmt.Println(objPkgName)
	fmt.Println(objPkgName.Imported().Scope().Lookup("Errorf"))
}

var src = `package p

import fmtPkg "fmt"

func main() {
	fmtPkg.Println("Hello, world")
}
`
