# Simplify: コードレビュー結果と修正計画

## Context

`/simplify` による変更コードのレビュー。3つの観点（再利用性、品質、効率性）でレビューした結果、修正すべき問題を特定した。

## 修正する項目

### 1. Dead field: `Handler.w` の削除
- **File:** `internal/diagnostic/diagnostic.go`
- `Handler` 構造体の `w io.Writer` フィールドが `enc` 作成後一度も使われていない
- `w` フィールを削除する

### 2. `r.Attrs` コールバックの短絡評価
- **File:** `internal/diagnostic/diagnostic.go`
- `path` と `reason` の両方が見つかったら `return false` で早期終了する

### 3. `slog.DiscardHandler` の重複アロケーション回避
- **File:** `internal/discovery/discovery.go`
- `Scan` が呼ばれるたびに `slog.New(slog.DiscardHandler)` を生成している
- package-level の変数に変更する

### 4. Kind validation の重複排除
- **File:** `internal/metadata/metadata.go`
- `Validate` が `validKinds[m.Kind]` を直接チェックしており、`ParseKind` と同じロジックが重複
- 共通の `Kind.IsValid()` メソッドまたは内部ヘルパーに統一する

## 修正しない項目（理由付き）

- **CLI テストのボイラープレート重複**: テストコードのみの問題であり、現時点でスコープ外
- **`ListCmd.Run` の `os.Stdout`/`os.Stderr` 直接参照**: アーキテクチャ変更のため、スコープ外
- **`findGitRoot` の抽象化不足**: 同上
- **`processFile` の3タプル返値**: リファクタリングの範囲が広いためスコープ外
- **`KindUnknown` が valid として扱われる件**: 設計意図の確認が必要なため、今回はスキップ
- **discovery の逐次処理**: 現在のスケールでは問題なし
- **diagnostic 属性キーの定数化**: 変更箇所が多くスコープ外
- **full goldmark parse / full file read**: 現在のスケールでは問題なし
- **git subprocess**: 同上

## 検証

- `task check` (lint + build + test) を実行して全パス確認
