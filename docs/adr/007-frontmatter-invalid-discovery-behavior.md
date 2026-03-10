# ADR-007: Invalid frontmatter behavior for discovery commands

## Status

Accepted

## Context

frontmatter discovery では、構文エラーだけでなく semantic integrity の崩れも invalid として扱う必要がある。例として required field の欠落、未知 enum、`related` のリンク切れ、選別に影響する `canonical` の不整合がある。

このとき command 全体を常に失敗させると repo-wide discovery が運用しにくくなる。一方で best-effort で黙殺すると、壊れた metadata が routing signal として混ざる。

コマンド形態ごとに失敗の粒度を固定する必要がある。

## Decision

### Batch discovery commands

`pd list` や `pd related` のような batch discovery command は、不正文書があっても command 全体は継続する。

不正文書は document 単位で invalid として扱い、diagnostics に invalid reason を出す。黙って無視しない。

### Single-document command

`pd show <path>` は対象文書が invalid な場合、non-zero exit で失敗させる。失敗時は invalid reason を返す。

### Invalid discovery state

次は invalid discovery state として扱う。

- required field の欠落
- 許可されていない enum value
- `title` 不在かつ H1 不在
- `related` に存在しない path が含まれる状態
- selection の正しさを壊す `canonical` の競合

## Consequences

- repo-wide discovery では不正文書があっても有効文書の探索を継続できる
- 単一文書の検査では失敗が終了コードに反映されるため、CLI 利用者が invalid を見落としにくい
- implementation は batch と single-target で error handling を分ける必要がある
- diagnostics の wire shape は将来調整できるが、invalid を黙殺しないという契約は固定される
