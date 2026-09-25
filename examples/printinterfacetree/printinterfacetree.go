package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go/types"

	"golang.org/x/tools/go/packages"
)

var (
	modeJSON   bool
	modeSimple bool
)

func main() {
	progname := filepath.Base(os.Args[0])

	flag.BoolVar(&modeJSON, "json", false, "Output JSON")
	flag.BoolVar(&modeSimple, "s", false, "Use simple names ('Node' instead of 'go/ast.Node')")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <pkg> <name>\n", progname)
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(2)
	}

	path := flag.Arg(0)
	name := flag.Arg(1)

	log.SetPrefix(progname + ": ")
	log.SetFlags(0)

	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedName | packages.NeedTypes}, path)
	if err != nil {
		log.Fatal(err)
	}
	if packages.PrintErrors(pkgs) > 0 || len(pkgs) != 1 {
		log.Fatalf("could not load package %q", path)
	}

	pkg := pkgs[0].Types
	pkgScope := pkg.Scope()

	rootObj := pkgScope.Lookup(name)
	if rootObj == nil {
		log.Fatalf("could not find %q in package %q", name, pkg)
	}

	root := rootObj.Type()
	if !types.IsInterface(root) {
		log.Fatalf("not an interface type: %s", types.ObjectString(rootObj, types.RelativeTo(pkg)))
	}

	isChild := map[types.Type]map[types.Type]bool{root: {}}

	for _, name := range pkgScope.Names() {
		obj, ok := pkgScope.Lookup(name).(*types.TypeName)
		if !ok {
			continue
		}
		if obj == rootObj {
			continue
		}
		if !obj.Exported() {
			continue
		}

		typ := obj.Type()

		if implementsInterface(typ, root) {
			isChild[root][typ] = true
		} else if ptyp := types.NewPointer(typ); implementsInterface(ptyp, root) {
			isChild[root][ptyp] = true
		}
	}

	detachNodes(root, isChild)

	if modeJSON {
		printJSON(root, isChild)
	} else {
		printPlain(root, isChild, 0)
	}
}

type typesSlice []types.Type

func (s typesSlice) Less(i, j int) bool {
	iSlice := types.IsInterface(s[i])
	jSlice := types.IsInterface(s[j])

	if iSlice != jSlice {
		return iSlice
	}

	return s[i].String() < s[j].String()
}

func (s typesSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s typesSlice) Len() int { return len(s) }

func sortedTypes(tree map[types.Type]bool) []types.Type {
	typs := []types.Type{}
	for c := range tree {
		typs = append(typs, c)
	}

	sort.Sort(typesSlice(typs))

	return typs
}

type stringNode struct {
	Name     string
	Children []*stringNode `json:",omitempty"`
}

func typeString(typ types.Type) string {
	if modeSimple {
		return types.TypeString(typ, func(_ *types.Package) string { return "" })
	} else {
		return types.TypeString(typ, nil)
	}
}

func printPlain(node types.Type, isChild map[types.Type]map[types.Type]bool, depth int) {
	fmt.Printf("%s%s\n", strings.Repeat("  ", depth), typeString(node))

	for _, c := range sortedTypes(isChild[node]) {
		printPlain(c, isChild, depth+1)
	}
}

func printJSON(node types.Type, isChild map[types.Type]map[types.Type]bool) {
	sNode := buildStringNode(node, isChild)
	json.NewEncoder(os.Stdout).Encode(sNode)
}

func buildStringNode(node types.Type, isChild map[types.Type]map[types.Type]bool) *stringNode {
	var sNode stringNode
	sNode.Name = typeString(node)
	sNode.Children = []*stringNode{}
	for _, child := range sortedTypes(isChild[node]) {
		sChild := buildStringNode(child, isChild)
		sNode.Children = append(sNode.Children, sChild)
	}
	return &sNode
}

func implementsInterface(typ types.Type, maybeInterface types.Type) bool {
	iface, ok := maybeInterface.Underlying().(*types.Interface)
	if !ok {
		return false
	}

	return types.Implements(typ, iface)
}

func detachNodes(parent types.Type, isChild map[types.Type]map[types.Type]bool) {
	newChildren := map[types.Type]bool{}

	for node := range isChild[parent] {
		foundNewParent := false

		for sibling := range isChild[parent] {
			if sibling == node {
				continue
			}
			if _, ok := isChild[node][sibling]; ok {
				// we have already seen that sibling is a child of node;
				// otherwise could not happen.
				continue
			}

			if implementsInterface(node, sibling) {
				// log.Printf("found implements: %s -> %s", node, sibling)
				if isChild[sibling] == nil {
					isChild[sibling] = map[types.Type]bool{}
				}
				isChild[sibling][node] = true
				foundNewParent = true
			}
		}

		if !foundNewParent {
			newChildren[node] = true
		}
	}

	isChild[parent] = newChildren

	for node := range newChildren {
		detachNodes(node, isChild)
	}
}
