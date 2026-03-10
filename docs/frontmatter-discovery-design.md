# Design Doc: Frontmatter-Based Discovery Entry

## Goal

`docs/**/*.md` を、本文の偶然的な冒頭ではなく frontmatter metadata から安定 discovery できる入口を `pd` に持たせる。

このテーマは Agent Skills における Progressive Disclosure の考え方に強く影響を受けている。つまり、最初に軽量な案内信号だけを見て必要な対象へ段階的に進み、最初から全文や全件を読まない導線を設計原則として採る。

この設計で成立させることは次の 3 点に絞る。

- Markdown 本体を唯一の SSoT としたまま metadata を読む
- 本文を開く前に「読むべき文書か」を安定判定できる
- 候補文書の絞り込み後にだけ本文読みに進める

frontmatter は本文要約の置き場ではなく、LLM や agent が「この文書を読むべきか」を判定する discovery metadata として扱う。`pd` はその metadata を安定して読むための deterministic な入口であり、本文そのものの代替や外部 index ではない。

## Scope / Non-Scope

### Scope

- `docs/**/*.md` から frontmatter を読み取る discovery 入口
- `kind` / `status` / `canonical` / `tags` を使った metadata 一覧化と絞り込み
- 単一文書の metadata 表示と、必要時だけ本文表示へ進む二段階導線
- `related` をたどるための read-only discovery 入口
- malformed frontmatter や schema 不一致を fail-closed で扱う方針
- discovery metadata の最小フィールドを前提にした read path の固定

### Non-Scope

- frontmatter の書き戻しや部分更新
- curator workflow や review queue の自動化
- repo-wide validation ルールの完全設計
- 本文要約生成や ranking / vector search
- 外部 metadata store や手管理 index の導入

## Architecture / Responsibility

### Responsibility split

- Markdown file
  - 本文と frontmatter を同居させる唯一の SSoT
- discovery reader
  - `docs/**/*.md` を走査し、frontmatter を安定抽出する
- metadata selector
  - `kind` / `status` / `canonical` / `tags` / `related` を機械可読に返す
- body escalation
  - metadata で候補決定後にだけ本文表示へ進ませる
- caller
  - discovery 結果を使って読む文書を選ぶ。本文の意味解釈はここより外側で行う

### Discovery metadata contract

- required fields
  - `kind`
  - `description`
  - `status`
- optional fields
  - `title`
  - `canonical`
  - `read_when`
  - `not_for`
  - `tags`
  - `related`
  - `metadata_reviewed_at`

このテーマで reader が依存してよいのは上記の frontmatter だけとする。唯一の例外として、`title` が未指定の場合に限り本文の最初の H1 を display identity として取得する。`title` も H1 もない文書は invalid とする。ファイル名や本文冒頭の散文からの追加推論は行わない。

### Boundary rules

- discovery reader は本文内容を解釈して metadata を補完しない。ただし `title` 未指定時の H1 fallback は許可する
- selector は read-only に徹し、frontmatter を修正しない
- body escalation は本文を返せるが、discovery 判定を本文依存に戻さない
- `related` は参照候補の提示に使うが、正本判定は `canonical` と `status` を優先する

### Minimal interface

- `pd list --json`
- `pd list --kind <kind> --status <status> --json`
- `pd show <path> --json`
- `pd show <path> --body`
- `pd related <path> --json`

## ADR References

- [ADR-005](docs/adr/005-frontmatter-minimal-schema.md): frontmatter の最小必須 schema
- [ADR-006](docs/adr/006-frontmatter-discovery-semantics.md): `kind` / `status` / `canonical` の semantics
- [ADR-007](docs/adr/007-frontmatter-invalid-discovery-behavior.md): invalid frontmatter の command failure model
- [ADR-008](docs/adr/008-frontmatter-reader-technology.md): reader の parser / fallback 技術選定
- [ADR-009](docs/adr/009-frontmatter-cli-and-validation-technology.md): CLI / decode / validation 技術選定

## Constraints

- Markdown 本体以外を SSoT にしない
- discovery metadata は本文要約の重複にしない
- malformed frontmatter は fail-open にせず、valid document として扱わない。batch command はファイル単位で invalid とし継続、single-document command は non-zero exit で失敗させる
- `title` 不在時は H1 fallback のみ許可し、両方不在なら invalid とする。ファイル名や散文からの推論は禁止する
- `canonical` と `status` を無視して本文読みに誘導しない
- 本文を広く読む前に metadata を確認できる導線を壊さない
- 更新責務を混在させない。このテーマでは read-only 入口だけを固定する
- 実装は標準ライブラリを第一選択とし、非標準依存は ADR で固定された責務を直接満たす場合に限る
- helper concern のために convenience dependency を広げない。file walking、path handling、JSON、error modeling、filtering、validation orchestration は std を使う

## Verification Strategy

### Required verification coverage

- `static + integration`

### Static

- frontmatter schema を表す型と decode 経路が一致すること
- unsupported field shape を silent accept しないこと
- CLI 出力契約が機械可読 JSON として安定していること

### Integration

- valid frontmatter を持つ複数文書から metadata 一覧を取得できること
- `kind` / `status` / `canonical` / `tags` で絞り込みできること
- `pd show --json` では metadata のみ、`pd show --body` では本文まで進めること
- malformed frontmatter、missing required field、duplicate canonical を reject できること
- `related` が存在しない path を含むとき、discovery 信号として不正を返せること

### Gate intent

- `static`
  - schema drift、decode 漏れ、禁止 field shape の見逃しを reject する
- `integration`
  - discovery 不達、canonical 判定不安定、body escalation の責務混在を reject する

## TODO Input

この Design Doc から切り出せる縦テーマ候補は次の通り。

- metadata 一覧から active / canonical な候補文書を安定選別できる
- 単一文書の metadata を見てから本文読みに進める
- `related` を discovery 導線として追跡できる

`Now` に進める候補は、最小の `list` / `show` 系で discovery の妥当性を最速検証できるものを優先する。

## Gate Input

- `static`
  - frontmatter decode 契約
  - required field 欠落時の失敗条件
- `integration`
  - `list` / `show` / `related` のシナリオ
  - duplicate canonical、invalid status、broken related の reject 条件

required verification coverage の各 gate に対し、`executor/check identifier + case identifier + pass/fail + replay handle` を残せるテスト実行面を用意する。

## Execution Input

- 実装順は `reader -> selector -> show/body escalation -> related`
- 最初に失敗させる verification は malformed frontmatter と duplicate canonical の integration test
- stop condition は次をすべて満たした時点
  - metadata だけで読むべき文書を絞れる
  - 本文読みが `show --body` に明示分離されている
  - required verification coverage に `reject-now` と `need-evidence` が残らない
