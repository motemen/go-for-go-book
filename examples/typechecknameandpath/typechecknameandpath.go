package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
)

func main() {
	path := "cmd/cover"

	bPkg, _ := build.Import(path, "", 0)

	fset := token.NewFileSet()
	files := []*ast.File{}
	for _, name := range bPkg.GoFiles {
		f, _ := parser.ParseFile(fset, filepath.Join(bPkg.Dir, name), nil, parser.Mode(0))
		files = append(files, f)
	}

	conf := types.Config{Importer: importer.Default()}
	pkg, _ := conf.Check(path, fset, files, nil)
	fmt.Printf("path=%v name=%v\n", pkg.Path(), pkg.Name())
}
