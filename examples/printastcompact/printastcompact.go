// printastcompact は Go のソースファイルを解析し、トップレベルの宣言の構文木を
// ast.Print に似た形式で、本文に載せやすいよう短く表示する。
// 位置情報、nil のフィールド、-omit で指定したフィールドは表示せず、
// 短いノードは1行にまとめる。
// -source を指定すると、各宣言の上にそのソースコードをコメントとして付ける。
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strings"
)

const maxInline = 72

var (
	posType   = reflect.TypeFor[token.Pos]()
	tokenType = reflect.TypeFor[token.Token]()
	identType = reflect.TypeFor[*ast.Ident]()
)

var omitted = map[string]bool{}

func main() {
	omit := flag.String("omit", "Obj", "comma-separated field names to omit")
	showSource := flag.Bool("source", false, "print the source of each declaration as a comment")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: printastcompact [-omit Field,...] <file>\n")
		os.Exit(2)
	}

	for name := range strings.SplitSeq(*omit, ",") {
		omitted[name] = true
	}

	src, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, flag.Arg(0), src, parser.SkipObjectResolution)
	if err != nil {
		log.Fatal(err)
	}

	for i, decl := range f.Decls {
		if i > 0 {
			fmt.Println()
		}
		if *showSource {
			printSource(fset, src, decl)
		}
		fmt.Println(format(reflect.ValueOf(decl), ""))
	}
}

// printSource は decl のソースコードを "// " を付けて表示する。関数の本体は省く。
func printSource(fset *token.FileSet, src []byte, decl ast.Decl) {
	end := decl.End()
	if fd, ok := decl.(*ast.FuncDecl); ok && fd.Body != nil {
		end = fd.Body.Pos()
	}
	code := string(src[fset.Position(decl.Pos()).Offset:fset.Position(end).Offset])
	for line := range strings.SplitSeq(strings.TrimSpace(code), "\n") {
		fmt.Println("// " + line)
	}
}

// format は v を indent の深さで文字列にする。収まるなら1行に、そうでなければ複数行にする。
func format(v reflect.Value, indent string) string {
	switch v.Kind() {
	case reflect.Interface:
		return format(v.Elem(), indent)

	case reflect.Pointer:
		// *ast.Ident の Obj などを除けば Name だけが残る
		return v.Type().String() + " " + format(v.Elem(), indent)

	case reflect.Struct:
		var items []string
		for i := range v.NumField() {
			name, fv := v.Type().Field(i).Name, v.Field(i)
			if omitted[name] || fv.Type() == posType || isNil(fv) {
				continue
			}
			items = append(items, name+": "+format(fv, indent+"    "))
		}
		return block("{", "}", items, indent)

	case reflect.Slice:
		// []*ast.Ident は名前だけ並べる
		if v.Type().Elem() == identType {
			names := make([]string, v.Len())
			for i := range v.Len() {
				names[i] = v.Index(i).Interface().(*ast.Ident).Name
			}
			return "[" + strings.Join(names, ", ") + "]"
		}
		items := make([]string, v.Len())
		for i := range v.Len() {
			items[i] = format(v.Index(i), indent+"    ")
		}
		return block("[", "]", items, indent)

	case reflect.String:
		return fmt.Sprintf("%q", v.String())

	default:
		if v.Type() == tokenType {
			return v.Interface().(token.Token).String()
		}
		return fmt.Sprint(v.Interface())
	}
}

func block(open, close string, items []string, indent string) string {
	if len(items) == 0 {
		return open + close
	}
	inline := open + " " + strings.Join(items, ", ") + " " + close
	if len(indent)+4+len(inline) <= maxInline && !strings.Contains(inline, "\n") {
		return inline
	}
	var b strings.Builder
	b.WriteString(open + "\n")
	for _, item := range items {
		b.WriteString(indent + "    " + item + "\n")
	}
	b.WriteString(indent + close)
	return b.String()
}

func isNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Slice, reflect.Map:
		return v.IsNil()
	}
	return false
}
