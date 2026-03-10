# ADR-007: Invalid frontmatter behavior for discovery commands

## Status

Accepted

## Context

frontmatter discovery では、構文エラーだけでなく semantic integrity の崩れも invalid として扱う必要がある。例として required field の欠落、未知 enum、`title` 不在かつ H1 不在がある。

このとき command 全体を常に失敗させると repo-wide discovery が運用しにくくなる。一方で best-effort で黙殺すると、壊れた metadata が routing signal として混ざる。

コマンド形態ごとに失敗の粒度を固定する必要がある。

CLI 利用と機械処理の両方を安定させるため、error は JSON として返す必要がある。また `related` / `canonical` 起因の invalid は、それらの field をこの discovery scope に含めないため対象外とする。

## Decision

### Batch discovery commands

`pd list` のような batch discovery command は、不正文書があっても command 全体は継続する。

不正文書は document 単位で invalid として扱い、`stderr` に JSON で invalid reason を出す。黙って無視しない。

### Single-document command

`pd show <path>` は対象文書が invalid な場合や対象解決に失敗した場合、non-zero exit で失敗させる。失敗時は `stderr` に JSON で reason を返す。

### Invalid discovery state

次は invalid discovery state として扱う。

- required field の欠落
- 許可されていない enum value
- `title` 不在かつ H1 不在

## Consequences

- repo-wide discovery では不正文書があっても有効文書の探索を継続できる
- 単一文書の検査では失敗が終了コードに反映されるため、CLI 利用者が invalid を見落としにくい
- invalid と command failure の両方が `stderr` JSON になるため、human-readable text への依存を避けられる
- implementation は batch と single-target で error handling を分ける必要がある
- diagnostics は `stderr` JSON で返す契約を固定できる
