// Command goapi は、標準ライブラリの各シンボルが導入された Go のバージョンと、
// 非推奨になったバージョンを調べて JSON に書き出す。macro.rb はこの JSON を使って、
// since:/deprecated: マクロと godoc:: マクロにバージョンのバッジを付ける。
//
//	go run ./scripts/goapi -go 1.27.1 -repo .cache/golang-go > data/goapi.json
//
// -repo には github.com/golang/go のクローンを指定する（タグとオブジェクトが
// 取れればよいので `git clone --filter=tree:0 --no-checkout` で十分）。
//
// 導入バージョンは、指定したバージョンのタグにある api/go1.*.txt を古い順に読み、
// 最初に現れたファイルのバージョンとする。pkg.go.dev の「added in」と同じ考え方。
//
// 非推奨になったバージョンは、api ファイルの //deprecated の印では正確に
// わからない。印は Go 1.16 のころに過去の分がまとめて書き足されていて
// （たとえば go/importer.For は実際には Go 1.12 で非推奨になったが、印は
// go1.16.txt にある）、後のリリースで書き換えられることもあるため
// （go/ast.Object の印は Go 1.22 の開発中に go1.21.txt にも書き足された）。
// そこで、印の付いたシンボルのうち go/ 以下のパッケージのものについて、
// 各リリースタグのソースを go/parser で読み、ドキュメントコメントに
// 「Deprecated:」の段落が最初に現れたバージョンを調べる。
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Output は data/goapi.json の形式。
type Output struct {
	// Go は対象にした Go のバージョン。
	Go string `json:"go"`
	// Symbols は go doc 形式のシンボル名（例：go/ast.File.FileStart）から、
	// [導入されたマイナーバージョン] または
	// [導入されたマイナーバージョン, 非推奨になったマイナーバージョン] への対応。
	// Go 1.0 からあるものは導入バージョンが 0。
	// 非推奨になったバージョンは go/ 以下のパッケージについてだけ調べている。
	Symbols map[string][]int `json:"symbols"`
}

var (
	// "pkg go/ast, " や "pkg syscall (linux-386), " の部分
	rxPkg = regexp.MustCompile(`^pkg ([^ ,]+)(?: \([^)]*\))?, (.*)$`)
	// method (*File) Pos() token.Pos
	rxMethod = regexp.MustCompile(`^method \(\*?([A-Za-z0-9_]+)(?:\[[^\]]*\])?\) ([A-Za-z0-9_]+)`)
	// type File struct, FileStart token.Pos / type Node interface, End() token.Pos
	rxMember = regexp.MustCompile(`^type ([A-Za-z0-9_]+)(?:\[[^\]]*\])? (?:struct|interface), (?:embedded )?\*?(?:[A-Za-z0-9_]+\.)?([A-Za-z0-9_]+)`)
	// const X ... / var X ... / func X(... / type X ...
	rxDecl = regexp.MustCompile(`^(?:const|var|func|type) ([A-Za-z0-9_]+)`)

	rxDeprecated = regexp.MustCompile(`(?m)^Deprecated:`)
)

var repo string

func main() {
	goVersion := flag.String("go", "", "対象にする Go のバージョン（例：1.27.1）")
	flag.StringVar(&repo, "repo", "", "github.com/golang/go のクローンのパス")
	flag.Parse()
	if *goVersion == "" || repo == "" {
		flag.Usage()
		os.Exit(2)
	}
	latest, err := strconv.Atoi(strings.Split(*goVersion, ".")[1])
	if err != nil {
		log.Fatalf("bad Go version %q: %v", *goVersion, err)
	}

	out := Output{Go: *goVersion, Symbols: map[string][]int{}}
	marked := map[string]bool{} // api ファイルで //deprecated の印が付いているもの

	for v := 0; v <= latest; v++ {
		name := "api/go1.txt"
		if v > 0 {
			name = fmt.Sprintf("api/go1.%d.txt", v)
		}
		data, err := gitShow("go"+*goVersion, name)
		if err != nil {
			log.Fatal(err)
		}
		s := bufio.NewScanner(bytes.NewReader(data))
		for s.Scan() {
			key, deprecated, ok := parseLine(s.Text())
			if !ok {
				continue
			}
			if _, seen := out.Symbols[key]; !seen {
				out.Symbols[key] = []int{v}
			}
			if deprecated {
				marked[key] = true
			}
		}
	}

	var keys []string
	for key := range marked {
		if strings.HasPrefix(key, "go/") {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		since := out.Symbols[key][0]
		dep := -1
		for v := max(since, 1); v <= latest; v++ {
			if deprecatedAt(key, v) {
				dep = v
				break
			}
		}
		if dep < 0 {
			log.Printf("%s: //deprecated の印はあるが、ドキュメントに Deprecated: が見つからない", key)
			continue
		}
		out.Symbols[key] = []int{since, dep}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	if err := enc.Encode(out); err != nil {
		log.Fatal(err)
	}
}

// parseLine は api ファイルの1行を、go doc 形式のシンボル名と //deprecated の印に分解する。
func parseLine(line string) (key string, deprecated bool, ok bool) {
	m := rxPkg.FindStringSubmatch(line)
	if m == nil {
		return "", false, false
	}
	pkg, rest := m[1], m[2]

	if i := strings.Index(rest, " //deprecated"); i >= 0 {
		deprecated = true
		rest = rest[:i]
	}
	if i := strings.Index(rest, " #"); i >= 0 {
		rest = rest[:i]
	}

	if m := rxMethod.FindStringSubmatch(rest); m != nil {
		return pkg + "." + m[1] + "." + m[2], deprecated, true
	}
	if m := rxMember.FindStringSubmatch(rest); m != nil {
		return pkg + "." + m[1] + "." + m[2], deprecated, true
	}
	if m := rxDecl.FindStringSubmatch(rest); m != nil {
		return pkg + "." + m[1], deprecated, true
	}
	return "", false, false
}

// releaseTag は Go 1.<minor> の最初のリリースのタグ名を返す。
func releaseTag(minor int) string {
	if minor <= 20 {
		return fmt.Sprintf("go1.%d", minor)
	}
	// Go 1.21 からはリリースタグが go1.21.0 のようになった
	return fmt.Sprintf("go1.%d.0", minor)
}

func gitShow(tag, file string) ([]byte, error) {
	out, err := exec.Command("git", "-C", repo, "show", tag+":"+file).Output()
	if err != nil {
		return nil, fmt.Errorf("git show %s:%s: %w", tag, file, err)
	}
	return out, nil
}

type pkgKey struct {
	pkg   string
	minor int
}

var pkgCache = map[pkgKey][]*ast.File{}

// parsePackage は Go 1.<minor> のリリース時点の標準パッケージ pkg を構文解析する。
func parsePackage(pkg string, minor int) []*ast.File {
	k := pkgKey{pkg, minor}
	if files, ok := pkgCache[k]; ok {
		return files
	}
	tag := releaseTag(minor)
	dir := "src/" + pkg
	if minor <= 3 {
		dir = "src/pkg/" + pkg // Go 1.4 より前は src/pkg 以下にあった
	}
	out, err := exec.Command("git", "-C", repo, "ls-tree", "--name-only", tag, dir+"/").Output()
	if err != nil {
		log.Fatalf("git ls-tree %s %s: %v", tag, dir, err)
	}
	var files []*ast.File
	fset := token.NewFileSet()
	for _, name := range strings.Fields(string(out)) {
		if path.Ext(name) != ".go" || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := gitShow(tag, name)
		if err != nil {
			log.Fatal(err)
		}
		f, err := parser.ParseFile(fset, name, src, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			log.Fatal(err)
		}
		if f.Name.Name == path.Base(pkg) {
			files = append(files, f)
		}
	}
	pkgCache[k] = files
	return files
}

// deprecatedAt は、シンボル key のドキュメントコメントに、Go 1.<minor> の
// リリース時点で Deprecated: の段落があるかを調べる。
func deprecatedAt(key string, minor int) bool {
	i := strings.LastIndex(key, "/")
	pkg, rest, _ := strings.Cut(key[i+1:], ".")
	pkg = key[:i+1] + pkg
	name, member, _ := strings.Cut(rest, ".")

	for _, f := range parsePackage(pkg, minor) {
		for _, decl := range f.Decls {
			for _, doc := range docsOf(decl, name, member) {
				if doc != nil && rxDeprecated.MatchString(doc.Text()) {
					return true
				}
			}
		}
	}
	return false
}

// docsOf は、宣言 decl がシンボル name（member があればその型のメソッドかフィールド）を
// 宣言しているとき、そのドキュメントコメントを返す。
func docsOf(decl ast.Decl, name, member string) []*ast.CommentGroup {
	switch decl := decl.(type) {
	case *ast.FuncDecl:
		if member == "" {
			if decl.Recv == nil && decl.Name.Name == name {
				return []*ast.CommentGroup{decl.Doc}
			}
			return nil
		}
		if decl.Recv != nil && decl.Name.Name == member && recvName(decl.Recv.List[0].Type) == name {
			return []*ast.CommentGroup{decl.Doc}
		}
	case *ast.GenDecl:
		for _, spec := range decl.Specs {
			switch spec := spec.(type) {
			case *ast.ValueSpec:
				for _, id := range spec.Names {
					if member == "" && id.Name == name {
						// const や var のグループ全体に付いたコメントも、その中の各定数・変数のもの
						return []*ast.CommentGroup{decl.Doc, spec.Doc, spec.Comment}
					}
				}
			case *ast.TypeSpec:
				if spec.Name.Name != name {
					continue
				}
				if member == "" {
					docs := []*ast.CommentGroup{spec.Doc, spec.Comment}
					if decl.Lparen == token.NoPos {
						docs = append(docs, decl.Doc)
					}
					return docs
				}
				var fields *ast.FieldList
				switch t := spec.Type.(type) {
				case *ast.StructType:
					fields = t.Fields
				case *ast.InterfaceType:
					fields = t.Methods
				}
				if fields == nil {
					return nil
				}
				for _, field := range fields.List {
					for _, id := range field.Names {
						if id.Name == member {
							return []*ast.CommentGroup{field.Doc, field.Comment}
						}
					}
					if len(field.Names) == 0 && recvName(field.Type) == member {
						return []*ast.CommentGroup{field.Doc, field.Comment}
					}
				}
			}
		}
	}
	return nil
}

// recvName は、レシーバや埋め込みフィールドの型の式から型名を取り出す。
func recvName(expr ast.Expr) string {
	for {
		switch e := expr.(type) {
		case *ast.StarExpr:
			expr = e.X
		case *ast.IndexExpr:
			expr = e.X
		case *ast.IndexListExpr:
			expr = e.X
		case *ast.SelectorExpr:
			return e.Sel.Name
		case *ast.Ident:
			return e.Name
		default:
			return ""
		}
	}
}
