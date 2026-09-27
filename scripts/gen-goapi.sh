#!/bin/sh
# data/goapi.json（標準ライブラリの各シンボルの導入・非推奨バージョン）を作り直す。
# config.json の Go のバージョンを上げたら実行する。
set -e

cd "$(dirname "$0")/.."

repo=.cache/golang-go
if [ ! -d "$repo" ]; then
  # タグとオブジェクトさえ取れればよいので、ツリーやファイルは必要になったときに取得する
  git clone --filter=tree:0 --no-checkout https://github.com/golang/go "$repo"
else
  git -C "$repo" fetch --tags --quiet
fi

go_version=$(ruby -rjson -e 'puts JSON.parse(File.read("config.json"))["Versions"]["Go"]')
go run ./scripts/goapi -go "$go_version" -repo "$repo" > data/goapi.json.tmp
mv data/goapi.json.tmp data/goapi.json
