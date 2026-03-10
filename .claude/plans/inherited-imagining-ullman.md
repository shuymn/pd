# Simplify: `--depth` オプション追加コードのレビュー結果と修正計画

## Context

`--depth` オプションを `pd list` に追加する変更に対して、コード再利用・品質・効率の3観点でレビューを実施。以下の修正が必要。

## 修正項目

### 1. `relativeDepth()` を `splitRelativePath()` で再利用 (コード再利用)

- **場所:** `internal/discovery/discovery.go:212-228`
- **問題:** `relativeDepth()` は `filepath.Clean` → `"."` チェック → セパレータ分割という処理を行うが、同パッケージの `splitRelativePath()` (`internal/discovery/gitignore.go:268-275`) が同じロジックを持つ
- **修正:** `relativeDepth()` を `splitRelativePath()` を使って書き直す:
  ```go
  func relativeDepth(path string, isDir bool) int {
      dir := path
      if !isDir {
          dir = filepath.Dir(filepath.Clean(path))
      }
      return len(splitRelativePath(dir))
  }
  ```
- これにより `strings.Split` による不要なアロケーションも解消される（効率レビュー指摘 #3 も同時に解決）

### 2. ファイルレベルの `exceedsMaxDepth` チェックを削除 (デッドコード)

- **場所:** `internal/discovery/discovery.go:188-190`
- **問題:** `handleDir()` が深すぎるディレクトリに `fs.SkipDir` を返すため、その配下のファイルは `WalkDir` によって走査されない。ファイルレベルの depth チェックは到達不能コード
- **修正:** 188-190行の `exceedsMaxDepth` 呼び出しを削除

## 修正不要（スキップ）

- **`Depth` を `Root` から `ListCmd` に移動:** README で `pd show --depth 1` が例示され、設計ドキュメントで `show` は `--depth` を受け入れるが結果を変えないと明記。`Root` に置くのは意図的
- **`newWalkFunc` の6パラメータ:** 現時点では各パラメータが明確な役割を持ち、グルーピングの必要性は低い
- **CLI テストと discovery テストの重複:** 異なるテスト境界（E2E vs ユニット）をカバーしており意図的
- **depth バリデーションの抽象化:** 1箇所のみの使用で抽象化は過剰

## 対象ファイル

- `internal/discovery/discovery.go` - `relativeDepth` 書き直し、ファイルレベル depth チェック削除

## 検証

1. `task check` (lint + build + test) が全てパス
