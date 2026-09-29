package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"strings"
)

func main() {
	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "example.go", src, parser.Mode(0))

	lookup := func(path string) (io.ReadCloser, error) {
		fmt.Println("lookup:", path)
		out, err := exec.Command("go", "list", "-export", "-f", "{{.Export}}", path).Output()
		if err != nil {
			return nil, err
		}
		return os.Open(strings.TrimSpace(string(out)))
	}
	conf := types.Config{Importer: importer.ForCompiler(fset, "gc", lookup)}

	pkg, _ := conf.Check("path/to/pkg", fset, []*ast.File{f}, nil)
	fmt.Println(pkg.Scope().Lookup("conf").Type())
	fmt.Println(pkg.Scope().Lookup("fset").Type())
}

var src = `package p

import "golang.org/x/tools/go/packages"

var conf packages.Config
var fset = conf.Fset
`
