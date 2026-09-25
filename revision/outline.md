# 改訂版アウトライン（たたき台）

`tasks.csv` の G-04 のための叩き台です。現行版の章構成と、現代の Go 静的解析の「道具の積み重なり」に沿って並べ替えた案を並べています。確定したら index.adoc に反映してください。

## 現行版（2018年、Go 1.7.4 基準）

```
はじめに
1. 構文解析              (01-ast.adoc, 01-ast/*)        … ほぼ完成、ただし Object/Scope 部分は非推奨 API
2. コメントとドキュメント (02-comment-and-document.adoc) … 後半 WIP
3. ソースコードの文字列化 (WIP: go/printer, go/format)
4. 型解析                (04-types.adoc)                … 後半 TBD/WIP
5. ビルド情報            (WIP: go/build)
6. 高レベルのAPI         (WIP: tools/go/loader)          … loader は非推奨
7. ソースコードを読む     (07-reading-code*)              … gddo/guru は廃止済み
8. フロー解析            (WIP: tools/go/ssa)
```

## 改訂案

基本の考え方は、読者が最終的に「自作の解析ツールを go vet / gopls / go fix の仕組みに載せられる」ところまで到達できるように、下の層から順に積み上げることです。現行版は標準パッケージだけで閉じていましたが、現代の実用的なツールはほぼ全て `golang.org/x/tools` の go/packages と go/analysis の上に構築されているため、そこまでを本書の射程に入れる案にしています（G-02 で判断）。

```
はじめに
  - なぜ「GoのためのGo」か（go vet, gopls, go fix, //go:fix inline）
  - パッケージの全体地図: token → parser → ast → types → packages → analysis → ssa
  - 対象バージョンと表記ルール

第I部 構文
  1. 構文解析（go/token, go/parser, go/ast）
     式 / ファイル / 構文ノードの実装（ジェネリクス関連ノードを含む）
     構文木の走査（Walk, Inspect, Preorder, PreorderStack）
     ソースコード中の位置（token.File, //line）
     コラム: ast.Object と構文レベルの名前解決（非推奨の経緯）
  2. コメント・ディレクティブ・ドキュメント
     コメントの解析 / ディレクティブ（ast.ParseDirective）/ 生成コード判定
     go/doc と go/doc/comment / Examples
  3. ソースコードの出力と書き換え（go/printer, go/format, astutil.Apply）
     コメント位置ずれ問題 / テキスト編集ベースの書き換え

第II部 意味
  4. 型解析（go/types）
     型チェックの実行 / エラー / Info / Object / Type
     ジェネリクス（TypeParam, Instance, 型推論）/ Alias
     言語バージョン（Config.GoVersion, FileVersions, go/version）
  5. パッケージの読み込み（go/packages, go/build/constraint）
     LoadMode / テスト / Overlay / 仕組み（go list, GOPACKAGESDRIVER）
     エクスポートデータ

第III部 ツール
  6. 静的解析ツールを作る（go/analysis）
     Analyzer と Pass / inspect と Cursor / SuggestedFix / Facts
     analysistest / ドライバ（singlechecker, unitchecker, go vet -vettool）
     配布（gopls, golangci-lint, go fix）と //go:fix inline
  7. フロー解析（go/cfg, go/ssa, callgraph）

第IV部 事例
  8. ソースコードを読む（2〜3本に絞る）
     候補: stringer / gofmt -r / go fix のモダナイザ / go vet の printf / govulncheck

付録
  ast.Node の階層 / types.Object・types.Type の階層（自動生成）
  用語集
```

## 構成上の論点

事例章は、現行版のように最後にまとめる形と、各解説章の直後に「読んでみよう」として差し込む形のどちらもありえます（R-01）。後者は解説と実例が近くなる一方、事例が複数の章の知識を要求する場合（goimports など）に置き場所に困ります。

ast.Object 周り（現行 01-ast/scope.adoc）は、現行版では章の大きな部分を占めていますが、Go 1.22 で非推奨になりました。構文だけでは名前が解決できないことを示す導入としての価値はあるので、短いコラムに縮めて型解析への橋渡しに使う案にしています（A-10）。
