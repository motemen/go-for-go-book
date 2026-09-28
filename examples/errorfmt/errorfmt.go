// errorfmt は、errors.New(fmt.Sprintf(...)) という呼び出しを見つけて報告する。
package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
	"golang.org/x/tools/go/types/typeutil"
)

var Analyzer = &analysis.Analyzer{
	Name: "errorfmt",
	Doc:  "report errors.New(fmt.Sprintf(...)) that can be fmt.Errorf(...)",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for n := range ast.Preorder(file) {
			call, ok := n.(*ast.CallExpr)
			if !ok || !isFunc(pass.TypesInfo, call, "errors.New") {
				continue
			}
			if len(call.Args) != 1 {
				continue
			}
			arg, ok := call.Args[0].(*ast.CallExpr)
			if !ok || !isFunc(pass.TypesInfo, arg, "fmt.Sprintf") {
				continue
			}
			pass.Reportf(call.Pos(), "use fmt.Errorf(...) instead of errors.New(fmt.Sprintf(...))")
		}
	}
	return nil, nil
}

// isFunc は、call が name という名前の関数の呼び出しかどうかを返す。
func isFunc(info *types.Info, call *ast.CallExpr, name string) bool {
	fn, ok := typeutil.Callee(info, call).(*types.Func)
	return ok && fn.FullName() == name
}

func main() {
	singlechecker.Main(Analyzer)
}
