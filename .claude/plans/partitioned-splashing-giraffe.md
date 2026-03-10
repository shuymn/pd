# エラーメッセージ "unknown kind" → "invalid kind" へのリネーム

## Context

`KindUnknown`（値: `"unknown"`）は正当な Kind 定数だが、不正な Kind 値に対するエラーメッセージにも `"unknown kind:"` を使っているため紛らわしい。エラーメッセージを `"invalid kind:"` に統一して曖昧さを解消する。

## 変更対象

### 1. `internal/metadata/metadata.go`
- L34: `"unknown kind: %q"` → `"invalid kind: %q"`（`ParseKind`）
- L64: `"unknown kind: %q"` → `"invalid kind: %q"`（`Validate`）

### 2. `internal/metadata/metadata_test.go`
- L76: テスト名 `"unknown kind"` → `"invalid kind"`
- L78: `wantReason` を `"unknown kind: \"blog\""` → `"invalid kind: \"blog\""`

### 3. `TODO.md`（任意）
- L16, L29: `unknown kind` の記述を `invalid kind` に更新

## 対象外
- `KindUnknown` 定数自体はそのまま（既定値としての役割は変わらない）
- `docs/` 配下や `.claude/plans/` 配下の参照は変更しない

## 検証
- `go test ./internal/metadata/...` が全件パス
- `task check` が成功
