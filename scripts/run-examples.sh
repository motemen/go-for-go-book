#!/bin/sh
# examples/ 以下の main パッケージをすべて実行し、失敗したものがあれば非ゼロで終了する。
set -e
cd "$(dirname "$0")/.."

status=0
for dir in $(go list -f '{{if eq .Name "main"}}{{.Dir}}{{end}}' ./examples/...); do
  name=$(basename "$dir")
  # printinterfacetree は引数が必要なツールなので scripts/gen-listings.sh で確かめる
  [ "$name" = printinterfacetree ] && continue
  if go run "$dir" > /dev/null 2>&1; then
    echo "ok   $name"
  else
    echo "FAIL $name"
    go run "$dir" 2>&1 | sed 's/^/     /'
    status=1
  fi
done
exit $status
