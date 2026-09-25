#!/bin/sh
# 付録の階層リスト（listings/*.txt）を、使用中の Go のバージョンで生成し直す。
set -e
cd "$(dirname "$0")/.."

go run ./examples/printinterfacetree -s go/ast Node > listings/ast-node-hierarchy.txt
go run ./examples/printinterfacetree go/types Object > listings/types-object-hierarchy.txt
go run ./examples/printinterfacetree go/types Type > listings/types-type-hierarchy.txt
