# ADR-006: Frontmatter discovery semantics

## Status

Accepted

## Context

`pd list` や `pd show` を安定した discovery 入口にするには、frontmatter field の語彙を固定する必要がある。

特に `kind` は selection と filtering の中心であり、free-form にすると metadata が discovery signal として機能しない。逆に field を増やしすぎると導入負荷が過大になる。

`canonical` を導入すると selection semantics より先に corpus-wide uniqueness と validation 責務を背負い込むため、この ADR では主文書判定を固定しない。

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

## Consequences

- `pd list --kind ...` のフィルタ契約を validation と同じ語彙で固定できる
- `kind` 欠落や未知の enum は invalid として扱える
- `status` など別の routing signal をこの段階で持ち込まないため、discovery v1 の導入負荷を抑えられる
- `canonical` 主文書選別はこの ADR の対象外とし、主文書判定や重複解決は別 ADR で固定する
