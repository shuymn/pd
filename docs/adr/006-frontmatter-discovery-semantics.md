# ADR-006: Frontmatter discovery semantics

## Status

Accepted

## Context

`pd list` や `pd show` を安定した discovery 入口にするには、frontmatter field の語彙と既定値を固定する必要がある。

特に `kind` / `status` / `canonical` は selection と filtering の中心であり、free-form にすると metadata が discovery signal として機能しない。逆に全 field を厳格必須にすると導入負荷が過大になる。

## Decision

### `kind`

`kind` は required field とし、allowed values は次に固定する。

- `roadmap`
- `design-doc`
- `adr`
- `coding`
- `testing`
- `tooling`
- `review`
- `unknown`

`unknown` は read 時のエラーバケットではなく、明示的に許可された値とする。

### `status`

`status` は required field とし、allowed values は次に固定する。

- `active`
- `deprecated`
- `draft`

`pd list` は `--status` を指定しない場合、`active` のみを返す。`draft` と `deprecated` は明示指定時に返す。

### `canonical`

`canonical` は optional field とする。値がない場合は `false` とみなす。

- `canonical: true`
  - 同一または近接テーマの複数文書があるとき、優先して案内すべき主文書
- `canonical: false` または未記入
  - 主文書としては明示されていない文書

## Consequences

- `pd list --kind ...` と `pd list --status ...` のフィルタ契約を validation と同じ語彙で固定できる
- `status` 欠落や未知の enum は invalid として扱える
- 通常の discovery は `active` のみを見るため、deprecated / draft 文書のノイズが減る
- `canonical` は段階導入しやすいが、主文書選別を使う箇所では `true` を明示しない限り非 canonical と扱われる
