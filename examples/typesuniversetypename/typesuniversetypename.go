package main

import (
	"fmt"
	"go/types"
)

func main() {
	scope := types.Universe
	for _, name := range scope.Names() {
		obj, ok := scope.Lookup(name).(*types.TypeName)
		if !ok {
			continue
		}
		underlying := ""
		if obj.Type() != obj.Type().Underlying() {
			underlying = obj.Type().Underlying().String()
		}
		fmt.Printf("%-12s %-16T IsAlias=%-8v %s\n",
			name, obj.Type(), obj.IsAlias(), underlying)
	}
}
