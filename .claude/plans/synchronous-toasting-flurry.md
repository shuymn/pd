# metadata.Validate の reason 構造化

## Context

`discovery.go` の `processFile` で `metadata.Validate` のエラーを `err.Error()` で reason 文字列に変換している。現状 `Validate` はバリデーション理由（ユーザー向け診断）のみ返すが、将来的に内部エラー（例: 外部リソース参照）を返す必要が生じた場合、reason と内部エラーを区別できず破綻する。

## Approach

`Validate` のシグネチャを `error` から `(reason string, err error)` に変更する。

**根拠:** `processFile` 自体が `(*metadata.Result, string, error)` という `(値, reason, error)` 三つ組を既に使っており、同じ規約に揃える。カスタムエラー型（`ValidationError`）は現時点で内部エラーパスが存在しないため過剰設計。

## Steps

### 1. `internal/metadata/metadata.go` — シグネチャ変更

`Validate` を `(reason string, err error)` に変更:
- `return errors.New("...")` → `return "...", nil`
- `return fmt.Errorf(...)` → `return fmt.Sprintf(...), nil`
- `return nil` → `return "", nil`
- `"errors"` import を削除

### 2. `internal/discovery/discovery.go` — 呼び出し元更新 (L122-124)

```go
// Before
if err := metadata.Validate(meta); err != nil {
    return nil, err.Error(), nil
}

// After
reason, err := metadata.Validate(meta)
if err != nil {
    return nil, "", err
}
if reason != "" {
    return nil, reason, nil
}
```

内部エラーは第3返り値として伝播し、walk を中断させる。バリデーション理由は従来通り diagnostic logger へ。

### 3. `internal/metadata/metadata_test.go` — テスト更新

- `wantErr bool` + `errMsg string` → `wantReason string` に統一
- `err := metadata.Validate(...)` → `reason, err := metadata.Validate(...)`
- 全ケースで `err == nil` を検証（現時点で内部エラーパスなし）
- `reason` と `wantReason` を比較

### 4. `discovery_test.go` — 変更不要

既存テストは diagnostic JSON の `Reason` が非空かのみ検証しており、影響なし。

## Verification

- `task check` (lint + build + test) がパスすること
- `metadata.Validate` の呼び出し元は `discovery.go` L122 のみ（他に影響なし）
