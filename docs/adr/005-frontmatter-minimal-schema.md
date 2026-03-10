---
kind: adr
description: Read when you need the required frontmatter fields and schema contract for discovery.
---

# ADR-005: Frontmatter minimal schema for discovery

## Status

Accepted

## Context

`pd` の frontmatter discovery は、Markdown 本体を唯一の SSoT に保ったまま「読むべき文書か」を本文前に判定できることを目的とする。

導入コストを上げすぎず、Agent Skills の Progressive Disclosure に近い「最小の識別信号 + 最小の routing signal」から始めたい。一方で、field が少なすぎると discovery 入口としての妥当性が崩れる。

discovery field として `read_when` / `not_for` / `tags` / `related` / `metadata_reviewed_at` まで持ち込むと、read-only discovery 入口としては責務が広すぎる。そのため初期 schema はさらに絞る必要がある。

最小 schema では次の論点を固定する必要がある。

- discovery に必要な required field は何か
- `title` を必須にしない場合、識別子をどう決めるか
- 暗黙補完をどこまで許すか

## Decision

discovery 用 frontmatter の最小 schema は次とする。

- Required:
  - `kind`
  - `description`
- Optional:
  - `title`

`description` は本文要約ではなく、読むべき理由を示す discovery signal として扱う。

`title` は必須にしない。`title` がない場合、discovery で使う表示名は本文の最初の Markdown H1 から取得する。`title` も H1 もない文書は discovery 上 invalid とする。

暗黙補完は H1 fallback のみ許可する。ファイル名、path、本文冒頭の散文、その他の metadata 推論は使わない。

`status` / `read_when` / `not_for` / `tags` / `related` / `metadata_reviewed_at` はこの schema に含めない。

## Consequences

- 既存文書へ frontmatter を導入するとき、最低限 `kind` / `description` を埋めれば discovery に参加できる
- `title` は任意のため、本文の見出し構造が display identity の一部になる
- discovery field 群を初期必須にしないため、導入負荷と schema drift の両方を抑えられる
- H1 fallback 以外の暗黙推論を禁止することで discovery 挙動は安定するが、見出しが弱い文書は invalid になりうる
- schema 実装では required field 欠落と H1 不在を invalid reason として明示する必要がある
