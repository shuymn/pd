# Simplify: コードレビュー結果と修正プラン

## Context

`pd` CLIツールの初期実装（frontmatter discovery機能）に対するコード品質・効率・再利用性レビュー。新規コード約1,500行が対象。

## 修正対象の指摘事項

### 1. [効率] `goldmark.DefaultParser()` が毎回新規割り当て（Medium）
- **ファイル:** `internal/heading/heading.go:15`
- **問題:** `ExtractH1`が呼ばれるたびに新しいパーサーを生成。パーサーは設定に関して無状態なのでパッケージレベルで再利用可能。
- **修正:** パッケージレベル変数 `var defaultParser = goldmark.DefaultParser()` に切り出し

### 2. [品質] `discovery.go`にハードコードされた`"list"`コマンド文字列（High）
- **ファイル:** `internal/discovery/discovery.go:113,121,130`
- **問題:** discovery packageがCLIコマンド名を知っている（leaky abstraction）
- **修正:** `Scanner`に`Command string`フィールドを追加。`processFile`に`command`引数を渡し、Diagnostic構築時に使用。`ListCmd.Run`で`Scanner{Root: root.Root, Command: "list"}`とする。

### 3. [品質] `ListCmd.JSON`が`required`だが分岐に未使用（Medium）
- **ファイル:** `internal/cli/list.go:16`
- **問題:** `JSON bool`は`required:""`だが、`Run`は常にJSON出力。フラグの値を確認する分岐がない。
- **修正:** `required:""`を外してoptionalにし、デフォルトをJSONにする（`default:"true"`）。将来の他フォーマット対応の余地を残す。

### 4. [再利用] 重複する`writeFile`テストヘルパー（Medium）
- **ファイル:** `internal/cli/cli_test.go:54-66`, `internal/discovery/discovery_test.go:13-25`
- **問題:** 同一の`writeFile`関数が2箇所に定義されている
- **修正:** `internal/testutil/testutil.go`等に共通ヘルパーとして切り出し

### 5. [効率] `json.NewEncoder.Encode`で簡素化可能（Low）
- **ファイル:** `internal/diagnostic/diagnostic.go:17-29`
- **問題:** `json.Marshal` + `fmt.Fprintf(w, "%s\n", data)`は`json.NewEncoder(w).Encode(d)`で置換可能（Encodeは自動的に改行を付加）
- **修正:** `json.NewEncoder`に置換

## 指摘済みだが修正しない事項

| 指摘 | 理由 |
|------|------|
| pointer-to-slice parameter sprawl in discovery.go | 構造としては妥当。Scanner内部stateへの移行は設計変更の規模が大きくスコープ外 |
| processFileの3値返却 | 現時点では呼び出し元が1つのみ。sum-typeラッパーは過剰 |
| os.Stdout/Stderr hardcoded in ListCmd | CLI統合テストで検証済み。DI化はスコープ外 |
| 並行ファイル処理 | docs/の規模が小さいため不要 |
| frontmatter.Extractが全ファイル読み込み | 実用上問題なし |
| JSON-line writeパターンの重複 | セマンティクスが異なる（JSONL vs JSON array）。共通化は過剰 |
| Kind validationの2箇所のmap lookup | 意図的な設計（parse vs validate） |

## 修正順序

1. `heading.go` — goldmarkパーサーをパッケージレベル変数に
2. `discovery.go` — ハードコードされた`"list"`を外部から注入
3. `list.go` — `JSON`フィールドの扱いを整理
4. テストヘルパー — 共通`writeFile`の抽出
5. `diagnostic.go` — `json.NewEncoder`に置換

## 検証

```bash
task check   # lint + build + test
```
