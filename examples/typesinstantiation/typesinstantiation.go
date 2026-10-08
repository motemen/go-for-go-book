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
	f, err := parser.ParseFile(fset, "example.go", src, parser.SkipObjectResolution)
	if err != nil {
		panic(err)
	}

	conf := types.Config{Importer: importer.Default()}
	info := &types.Info{
		Uses:      map[*ast.Ident]types.Object{},
		Instances: map[*ast.Ident]types.Instance{},
	}

	pkg, err := conf.Check("path/to/pkg", fset, []*ast.File{f}, info)
	if err != nil {
		panic(err)
	}

	// ジェネリックな型をインスタンス化したとき
	genInt := pkg.Scope().Lookup("genInt").Type().(*types.Named)   // Gen[int]
	genBool := pkg.Scope().Lookup("genBool").Type().(*types.Named) // Gen[bool]

	compare("Gen[_].FieldT",
		genInt.Underlying().(*types.Struct).Field(0),
		genBool.Underlying().(*types.Struct).Field(0),
	)

	compare("Gen[_].Meth",
		genInt.Method(0),
		genBool.Method(0),
	)

	compare("Gen[_].Meth(argT)",
		genInt.Method(0).Signature().Params().At(0),
		genBool.Method(0).Signature().Params().At(0),
	)

	// GenMeth の型パラメータ U。TypeName には Origin() がないので
	// メソッドの Origin() から辿る
	uInt := genInt.Method(1).Signature().TypeParams().At(0)
	uBool := genBool.Method(1).Signature().TypeParams().At(0)
	fmt.Printf("# %s\n%s vs %s\nsame=%-5t sameOrigin=%t\n",
		"Gen[_].GenMeth[U]",
		uInt.Obj(),
		uBool.Obj(),
		uInt.Obj() == uBool.Obj(),
		genInt.Method(1).Origin().Signature().TypeParams().At(uInt.Index()) ==
			genBool.Method(1).Origin().Signature().TypeParams().At(uBool.Index()),
	)

	// ジェネリックな関数やメソッドをインスタンス化したとき
	// （ソースコード中の出現順に集める）
	var idents []*ast.Ident
	for n := range ast.Preorder(f) {
		if ident, ok := n.(*ast.Ident); ok {
			if _, ok := info.Instances[ident].Type.(*types.Signature); ok {
				idents = append(idents, ident)
			}
		}
	}
	params := func(ident *ast.Ident) *types.Tuple {
		return info.Instances[ident].Type.(*types.Signature).Params()
	}

	compare("Fun[_]",
		info.Uses[idents[0]].(*types.Func), // Fun[int] の Fun
		info.Uses[idents[1]].(*types.Func), // Fun[bool] の Fun
	)

	compare("Fun[_](argT)",
		params(idents[0]).At(0), // Fun[int]
		params(idents[1]).At(0), // Fun[bool]
	)

	compare("genInt.GenMeth[_](argU)",
		params(idents[2]).At(0), // genInt.GenMeth[[]int]
		params(idents[3]).At(0), // genInt.GenMeth[IntSlice]
	)
}

// a と b が同じオブジェクトかどうかと、それぞれの Origin() が同じかどうかを印字する
func compare[O interface {
	comparable
	fmt.Stringer
	Origin() O
}](label string, a, b O) {
	fmt.Printf("# %s\n%s vs %s\nsame=%-5t sameOrigin=%t\n", label, a, b, a == b, a.Origin() == b.Origin())
}

var src = `package p

type Gen[T any] struct {
	FieldT T
}

func (Gen[T]) Meth(argT T) {}

func (Gen[T]) GenMeth[U ~[]T](argU U) {}

func Fun[T any](argT T) {}

type IntSlice []int

var genInt Gen[int]
var genBool Gen[bool]

var _ = Fun[int]
var _ = Fun[bool]

var _ = genInt.GenMeth[[]int]
var _ = genInt.GenMeth[IntSlice]
`
