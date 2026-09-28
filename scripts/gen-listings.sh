#!/bin/sh
# listings/ 以下の生成物（付録の階層リスト、構文木の出力）を、使用中の Go のバージョンで生成し直す。
set -e
cd "$(dirname "$0")/.."

go run ./examples/printinterfacetree -s go/ast Node > listings/ast-node-hierarchy.txt
go run ./examples/printinterfacetree go/types Object > listings/types-object-hierarchy.txt
go run ./examples/printinterfacetree go/types Type > listings/types-type-hierarchy.txt

go run ./examples/printastcompact -source -omit Obj,Body,Params,Results listings/typeparams-decl.go.txt > listings/typeparams-decl.ast.txt
go run ./examples/printastcompact -source -omit Obj,Body,Params,Results listings/typeparams-index.go.txt > listings/typeparams-index.ast.txt
go run ./examples/printastcompact -source -omit Obj,Incomplete listings/typeparams-interface.go.txt > listings/typeparams-interface.ast.txt
