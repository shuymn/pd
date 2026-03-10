---
kind: design-doc
description: Read when you need the full architecture and constraints for the frontmatter discovery entry.
---

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
- `kind` を使った metadata 一覧化と絞り込み
- 単一文書の metadata 表示と、必要時だけ本文表示へ進む二段階導線
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
  - discovery root 配下の Markdown を走査し、Git 管理下のディレクトリでは `.gitignore` と `.git/info/exclude` を尊重して frontmatter を安定抽出する
- metadata selector
  - `kind` を機械可読に返す
- body escalation
  - metadata で候補決定後にだけ本文表示へ進ませる
- caller
  - discovery 結果を使って読む文書を選ぶ。本文の意味解釈はここより外側で行う

### Discovery metadata contract

- required fields
  - `kind`
  - `description`
- optional fields
  - `title`

このテーマで reader が依存してよいのは上記の frontmatter だけとする。唯一の例外として、`title` が未指定の場合に限り本文の最初の H1 を display identity として取得する。`title` も H1 もない文書は invalid とする。ファイル名や本文冒頭の散文からの追加推論は行わない。

### Boundary rules

- discovery reader は本文内容を解釈して metadata を補完しない。ただし `title` 未指定時の H1 fallback は許可する
- selector は read-only に徹し、frontmatter を修正しない
- body escalation は本文を返せるが、discovery 判定を本文依存に戻さない
- 文書間の関連づけや正本判定の責務は持たない。必要なら Markdown 本文や将来の別テーマで扱う

### Minimal interface

- `pd list --json`
- `pd list --kind <kind> --json`
- `pd show <path> --json`
- `pd show <path> --body`

`pd` の path surface は常に discovery root 相対とする。`--root` 未指定時は discovery root が current working directory の `.`、`--root` 明示時は指定した subtree になる。`pd show` の入力 path、`pd list` / `pd show` の success metadata、diagnostics の `path` はすべてその discovery root 相対で扱う。
`pd list` の走査対象は Git 管理下のディレクトリでは `.gitignore` と `.git/info/exclude` を尊重する。一方で `pd show` は明示指定を優先し、ignore された path でも discovery root 内に存在すれば表示可能とする。

### Error contract

- `pd list` の invalid document は標準出力に混ぜず、通常実行では非表示にする
- `pd list --verbose` は invalid document を `stderr` に JSON で出す
- `pd show` の失敗理由は標準出力に混ぜず、`stderr` に JSON で出す
- JSON error は対象 path、machine-readable な reason を含む
- batch command は valid result を `stdout` に出して継続し、invalid ごとの error JSON は opt-in な `--verbose` でだけ `stderr` に出す
- single-document command は error JSON を `stderr` に出し、non-zero exit で失敗する

## ADR References

- [ADR-005](docs/adr/005-frontmatter-minimal-schema.md): frontmatter の最小必須 schema
- [ADR-006](docs/adr/006-frontmatter-discovery-semantics.md): `kind` の discovery semantics
- [ADR-007](docs/adr/007-frontmatter-invalid-discovery-behavior.md): invalid frontmatter の command failure model
- [ADR-008](docs/adr/008-frontmatter-reader-technology.md): reader の parser / fallback 技術選定
- [ADR-009](docs/adr/009-frontmatter-cli-and-validation-technology.md): CLI / decode / validation 技術選定

## Constraints

- Markdown 本体以外を SSoT にしない
- discovery metadata は本文要約の重複にしない
- malformed frontmatter は fail-open にせず、valid document として扱わない。batch command はファイル単位で invalid とし継続、single-document command は non-zero exit で失敗させる
- invalid や command failure の理由は human-only stderr text にせず、`stderr` JSON で機械可読に返す
- `title` 不在時は H1 fallback のみ許可し、両方不在なら invalid とする。ファイル名や散文からの推論は禁止する
- 本文を広く読む前に metadata を確認できる導線を壊さない
- 更新責務を混在させない。このテーマでは read-only 入口だけを固定する
- `canonical` と `related` はこのテーマで field として持ち込まない
- 実装は標準ライブラリを第一選択とし、非標準依存は ADR で固定された責務を直接満たす場合に限る
- helper concern のために convenience dependency を広げない。file walking、path handling、JSON、error modeling、filtering、validation orchestration は std を使う

## Verification Strategy

### Required verification coverage

- この discovery 入口では `static + integration` を required verification coverage とする
- `system` は、この設計が end-to-end の user flow 自体を直接扱う範囲に拡張されない限り要求しない

### Static

- frontmatter schema を表す型と decode 経路が一致すること
- unknown field や unsupported field shape を silent accept しないこと
- `kind` の語彙と validation 契約が一致すること

### Integration

- valid frontmatter を持つ複数文書から metadata 一覧を取得できること
- Git 管理下のディレクトリでは `.gitignore` や `.git/info/exclude` で除外された Markdown を `list` が列挙しないこと
- `kind` で絞り込みできること
- `pd show --json` では metadata のみ、`pd show --body` では本文まで進めること
- malformed frontmatter、missing required field、unknown field、unknown `kind`、`title` 不在かつ H1 不在、show 対象不在を reject できること
- `pd list --verbose` の invalid document と `pd show` failure reason が `stderr` JSON で観測できること
- CLI 出力契約が `stdout` success JSON / `pd list --verbose` のみ `stderr` diagnostic JSON / `pd show` failure `stderr` error JSON として安定していること

### Gate intent

- `static`
  - schema drift、decode 漏れ、禁止 field shape の見逃しを reject する
- `integration`
  - discovery 不達、error contract 不安定、body escalation の責務混在を reject する

## TODO Input

この Design Doc から評価対象として切り出せる縦テーマ候補は次の通り。

- metadata 一覧から `kind` ベースで候補文書を安定選別できる
- 単一文書の metadata を見てから本文読みに進める

`stderr` JSON の error contract は独立 Theme にはせず、各候補で `Reject if` と `Verification` を決めるための前提条件として扱う。

`Now` に進める候補は、最小の `list` / `show` 系で discovery の妥当性を最速検証できるものを優先する。

## Gate Input

- `static`
  - frontmatter decode 契約
  - required field 欠落、unknown field、unknown `kind`、`title` 不在かつ H1 不在の失敗条件
- `integration`
  - `list` / `show` のシナリオ
  - invalid frontmatter、missing required field、unknown field、unknown `kind`、`title` 不在かつ H1 不在、show 対象不在の reject 条件
  - `stderr` JSON emission の reject 条件

## Execution Input

- TODO 作成へ進む前に、候補ごとの required verification coverage を明示できること
- `list` 系と `show` 系の候補について、それぞれ `Outcome` と `Why not split further?` を後続で定義可能と判断できること
- stop condition は次をすべて満たした時点
  - `status` に依存する未確定判断が残っていない
  - `TODO Input` が縦テーマ候補として評価可能な粒度になっている
  - Gate owner ごとの reject 条件を後続で定義できる
  - `WORKFLOW.md` の Gate で `reject-now` と `need-evidence` が残らない
