---
kind: unknown
description: Read when you need to see pending feature themes and their acceptance criteria.
---

# TODO

- [ ] Theme: `pd list` の walk 範囲を `--root` 相対の `--depth` で制限できる
  - Outcome: グローバル `--depth` を指定すると、`pd list` は discovery root を深さ 0 として指定階層までだけ探索する。`--root=/` のような広い起点でも walk 範囲を明示的に抑制できる。`pd show` は明示 path 指定を優先し、`--depth` 指定の有無で成否が変わらない。
  - Why now: 現状は `--root` を広くすると `fs.WalkDir` が深く潜り続け、探索量と対象集合の両方が不必要に膨らむ。CLI 契約を大きく増やさずに探索境界だけを固定できる最小テーマとして切り出せる。
  - Verification: static + integration + system
  - Reject if:
    - [static] `--depth` 対応の責務が `list` 専用の探索境界から漏れ、共有 path 正規化や `show` の明示 path 解決に侵入する
    - [static] 深さ判定の基準が discovery root 相対の単一ロジックに集約されず、呼び出し経路ごとに別実装が増える
    - [static] required verification coverage の `static` gate に対する executor/check identifier、case/suite/scenario identifier、pass/fail、replay handle が記録されない
    - [integration] `--depth` の意味が discovery root 相対で固定されず、`--root` と組み合わせたときの深さ基準が曖昧なままになる
    - [integration] 深さ超過ディレクトリを剪定できず、指定深さを超える walk を継続してしまう
    - [integration] `pd list --depth 0 --json` が discovery root 直下の文書だけを安定して返さない
    - [integration] `pd list --depth 1 --json` が 1 階層下までの文書を返せない
    - [integration] `pd list --root <subtree> --depth 0 --json` で深さ基準が subtree 相対にならない
    - [integration] invalid な `--depth` 入力で非ゼロ終了にならず、負値や復元不能値を曖昧に受理する
    - [integration] depth 制限と kind filter を同時に指定したときに結果が壊れる
    - [integration] `.gitignore` で除外される path の扱いが depth 対応で回帰する
    - [integration] `pd show --depth ... <path>` が既存どおりの明示取得を維持できない
    - [integration] required verification coverage の `integration` gate に対する executor/check identifier、case/suite/scenario identifier、pass/fail、replay handle が記録されない
    - [system] agent-first な利用で広い `--root` に対して `--depth` を使った探索境界の抑制が効かず、必要な範囲だけを安全に列挙できない
    - [system] required verification coverage の `system` gate に対する executor/check identifier、case/suite/scenario identifier、pass/fail、replay handle が記録されない
  - Why not split further?: CLI フラグ追加、walk 剪定、`show` 非対象の境界は同じ探索契約を構成している。分けると「広い root でも安全に list できる」という外から観測可能な Outcome を一体で検証できない。

- [ ] Theme: H1 heading 抽出でインラインコードスパンのテキストが失われる
  - Outcome: `` `pd` / Frontmatter... `` のようにバッククォート囲みのコードスパンを含む H1 を持つ文書について、`pd list` / `pd show` のタイトルにコードスパン内のテキストが保持される（" / Frontmatter..." のような欠落が起きない）。
  - Why now: H1 fallback は `title` 未指定時の discovery identity として使われる。コードスパンが無言で消えると title が壊れたまま discovery に使われ、文書選択の信頼性が崩れる。
  - Verification: static + integration + system
  - Reject if:
    - [static] H1 抽出が goldmark の AST を歩かず、インラインノード種別（CodeSpan など）のテキストを無視したまま残る
    - [static] required verification coverage の `static` gate に対する executor/check identifier、case/suite/scenario identifier、pass/fail、replay handle が記録されない
    - [integration] バッククォート囲みのコードスパンを含む ATX H1 を持つ文書で `pd list` / `pd show` のタイトル表示にコードスパン内テキストが残らない
    - [integration] バッククォート囲みのコードスパンを含む Setext H1 を持つ文書で `pd list` / `pd show` のタイトル表示にコードスパン内テキストが残らない
    - [integration] H1 フォールバックの修正が `title` frontmatter 優先の既存 discovery/title 解決を回帰させる
    - [integration] required verification coverage の `integration` gate に対する executor/check identifier、case/suite/scenario identifier、pass/fail、replay handle が記録されない
    - [system] title 未指定文書を agent や利用者が一覧から選ぶ主要フローで、コードスパンを含む H1 のタイトル欠落により文書識別を誤る状態が残る
    - [system] required verification coverage の `system` gate に対する executor/check identifier、case/suite/scenario identifier、pass/fail、replay handle が記録されない
  - Why not split further?: goldmark AST のテキスト収集経路を修正する単一変更点であり、これ以上分割すると「コードスパンを含む H1 のタイトルが正しく返る」という Outcome を一体で検証できない。
