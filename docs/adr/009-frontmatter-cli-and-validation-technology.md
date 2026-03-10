# ADR-009: Frontmatter CLI and validation technology

## Status

Accepted

## Context

frontmatter discovery では `pd list` / `pd show` を実装し、frontmatter schema と invalid state を機械可読に扱う必要がある。

このとき固定すべき論点は次である。

- frontmatter decode を typed contract としてどう扱うか
- unknown field をどう reject するか
- semantic validation をどこで表現するか
- CLI command tree を何で実装するか

writer / validation / curation helper や `pd related` まで含めると CLI scope が広すぎるため、この ADR は read-only discovery 入口だけを固定対象にする。

ここでも標準ライブラリを第一選択とする。ただし、既に固定された責務を簡潔かつ明示的に実装するための依存は導入してよい。

## Decision

frontmatter の decode は typed Go struct を target に行う。動的 `map[string]any` を中間正本にしない。

semantic validation は decode 後に Go code で実装する。

- required field
- enum validation
- `title` 不在時の H1 fallback rule

unknown field は invalid discovery state として reject する。

CLI framework は `github.com/alecthomas/kong` を採用する。

YAML 専用ライブラリは `github.com/goccy/go-yaml` を採用する。`adrg/frontmatter` のカスタムフォーマット経由で strict decode を組み込む際、`gopkg.in/yaml.v2` の `UnmarshalStrict` は技術的には利用可能だが、次のポリシー判断により `goccy/go-yaml` を選択した。

- `yaml.v2` は legacy 扱いであり新規コードに持ち込む積極的理由がない
- `goccy/go-yaml` は `yaml.Strict()` オプションによる明示的な strict モード API を持ち、intent が読みやすい
- `goccy/go-yaml` はより活発にメンテナンスされている

`yaml.v2` は `adrg/frontmatter` の transitive dependency として `go.mod` に残るが、direct import は `goccy/go-yaml` に限定する。

CLI framework 以外の周辺処理は標準ライブラリを優先する。

- filesystem access
- JSON encoding
- slices/maps の操作
- error wrapping
- output formatting の基本処理

path scope は current working directory 基準で扱う。`--root` の default は `.` とし、`pd show` の `<path>`、`pd list` / `pd show` の success metadata、diagnostics の `path` はすべて選択された discovery root 相対で扱う。`../` による current working directory 外参照は reject し、absolute path は current working directory 配下の場合のみ許可する。

## Rejected alternatives

- `map[string]any` first で decode して後から整形する
  - schema drift を招きやすく、typed contract が弱いため不採用
- struct tag だけに validation を寄せる
  - cross-field rule と filesystem-backed rule を表しきれないため不採用
- external schema file を追加する
  - 管理対象を増やしすぎるため不採用
- `gopkg.in/yaml.v2` の `UnmarshalStrict` を使う
  - 技術的には機能するが legacy ライブラリであり、より明示的な API を持つ `goccy/go-yaml` を優先した
- std で足りる helper に convenience dependency を追加する
  - taste 由来の依存増加を避けるため不採用
- standard library のみで subcommand tree を組む
  - 今回の command tree と flag 契約では `kong` の方が責務を素直に表現できるため不採用

## Consequences

- implementation は typed decode と Go validation を前提に組める
- unknown field rejection のため追加 YAML dependency が本当に必要かは実装で見極める余地を残せる
- CLI は `kong` により subcommand / flag 定義を明示できる
- Git repository 探索を不要にし、current working directory 配下だけを見る単純な CLI 契約にできる
- `list` / `show` / diagnostics の path surface を常に discovery root 基準に揃えられ、機械処理の接続性を高められる
- writer / curation helper / `pd related` は実装対象に含めず、read-only discovery に責務を限定できる
- semantic validation の中心は `kind` と invalid discovery state になり、追加 routing field は別判断として切り離せる
- 非標準依存の追加は `adrg/frontmatter`、`goldmark`、`kong`、`goccy/go-yaml` に限定され、その他は std 優先で進める
