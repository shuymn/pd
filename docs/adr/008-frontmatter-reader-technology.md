---
kind: adr
description: Read when you need the technology selection for the frontmatter reader and H1 fallback.
---

# ADR-008: Frontmatter reader technology

## Status

Accepted

## Context

frontmatter discovery の reader は、Markdown 本体を唯一の SSoT に保ったまま metadata を先に読み、必要時だけ本文へ進む入口である。

この reader では次を満たす必要がある。

- frontmatter 境界を安定して抽出できる
- `title` 未指定時だけ本文の H1 を fallback として取得できる
- frontmatter decode と H1 fallback を同一責務に潰さない
- helper 実装に不要な依存を増やさない

repo の基本方針として、このテーマでは Go 標準ライブラリを第一選択とする。非標準依存は、既に設計で固定された責務を直接満たす場合に限って導入する。

## Decision

reader の技術選定は次に固定する。

- frontmatter 抽出
  - `github.com/adrg/frontmatter`
- H1 fallback 抽出
  - `github.com/yuin/goldmark`

reader は単一 parser pipeline に統一せず、責務ごとに分離する。

- frontmatter extraction / decode
- Markdown H1 fallback extraction

`title` fallback は `title` が未指定のときだけ実行する。fallback として認める見出しは Markdown の H1 として解釈されるものに限り、ATX H1 と Setext H1 を対象にする。

reader 周辺の補助処理は標準ライブラリを使う。

- file walking
- path handling
- JSON 出力
- error modeling
- filtering
- validation orchestration

## Rejected alternatives

- 単一の Markdown parser pipeline で frontmatter と H1 を同時に扱う
  - frontmatter extraction と H1 fallback の責務分離を崩すため不採用
- line-based custom parser
  - dependency は減るが、Markdown / frontmatter 境界の壊れ方に弱いため不採用
- filename/path fallback
  - 既存の schema 判断と衝突し、暗黙補完を増やすため不採用
- 汎用 helper のための追加ライブラリ導入
  - 標準ライブラリで十分な責務に convenience dependency を入れない

## Consequences

- reader 実装は `adrg/frontmatter` と `goldmark` だけを非標準依存として持つ
- H1 fallback の唯一の implicit derivation を parser ベースで安定化できる
- file/path/json/filter/error まわりは std 実装を前提に保てる
- 将来 reader 周辺に別ライブラリを足す場合は、新しい未充足責務があることを説明する追加 ADR が必要になる
