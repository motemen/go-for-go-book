#!/bin/sh
# examples/ 以下の main パッケージをすべて実行し、テストも実行する。失敗したものがあれば非ゼロで終了する。
set -e
cd "$(dirname "$0")/.."

status=0
# go/analysis の singlechecker・multichecker を使うサンプルは、引数が必要な解析ツールなので
# ここでは実行しない。本文の goexample マクロ（args）と、下の go test で確かめる。
for dir in $(go list -f '{{if eq .Name "main"}}{{.Dir}} {{.Imports}}{{end}}' ./examples/... | grep -v -e /singlechecker -e /multichecker | cut -d' ' -f1); do
  name=$(basename "$dir")
  # printinterfacetree と printastcompact は引数が必要なツールなので scripts/gen-listings.sh で確かめる
  [ "$name" = printinterfacetree ] && continue
  [ "$name" = printastcompact ] && continue
  if go run "$dir" > /dev/null 2>&1; then
    echo "ok   $name"
  else
    echo "FAIL $name"
    go run "$dir" 2>&1 | sed 's/^/     /'
    status=1
  fi
done

# テストのあるサンプル（analysistest を使う解析ツールなど）のテストを実行する
out=$(go test ./examples/... 2>&1) || status=1
echo "$out" | grep -v '\[no test files\]' || true
exit $status
