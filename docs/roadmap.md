---
kind: roadmap
description: Read when you need the current project roadmap and next planned themes.
---

# Roadmap

## Done

### Theme

Markdown repository を frontmatter ベースで安定 discovery できる入口を作る。

- Outcome: LLM が `docs/**/*.md` を本文の偶然的な冒頭ではなく metadata から絞り込み、読むべき文書を安定して選べる。
- Outcome note: 最初の固定対象は `kind` 中心の最小 schema に絞る。
- Why it matters: `ls` / `cat | head` 依存の探索だと文書種別、正本性、deprecated 文書の判別が不安定で、repo 内知識の利用精度が落ちる。
- Horizon: Done
- Promotion trigger: Markdown 本体を唯一の SSoT に保ったまま、最低限の frontmatter schema と metadata 読み取り体験を固定したい段階になったら `Design Doc` 化する。
- Delivered: `pd list --json` / `pd list --kind <kind> --json` で discovery metadata 一覧と kind 絞り込みを提供。`pd show <path> --json` / `--body` で単一文書の metadata と本文を段階的に開示。invalid document は stderr JSON で報告し non-zero exit で終了する。

## Now

### Theme

frontmatter を Markdown 本体に対して安全に保守できる更新経路を作る。

- Outcome: metadata 更新が free-form な YAML 再生成ではなく、構造化された変更として扱える。
- Why it matters: discovery metadata を育てても本文との整合性を壊しにくくなり、LLM や agent が更新時に壊れた frontmatter を作りにくくなる。
- Horizon: Now
- Promotion trigger: field の正規化、validation、部分更新の責務境界を固定しないと実装がぶれ始める段階になったら `Design Doc` 化する。

## Next

### Theme

本文執筆と metadata curation を分離した運用を作る。

- Outcome: 文書の著者が本文と core metadata を先に出し、後段の review / curator が discovery metadata を独立して改善できる。
- Why it matters: 執筆時の負荷を増やさずに corpus 全体の discovery 品質を継続改善できるようになる。
- Horizon: Next
- Promotion trigger: curator の責務、review queue、更新対象 field を固定しないと運用設計が定まらない段階になったら `Design Doc` 化する。

### Theme

frontmatter の整合性と repo-wide な語彙統制を検証できるようにする。

- Outcome: 必須 field 欠落、canonical の衝突、`related` のリンク切れ、タグ語彙の逸脱を継続的に検出できる。
- Why it matters: discovery metadata を導入しても corpus 全体で壊れたり発散したりすると routing 信号として信頼できない。
- Horizon: Next
- Promotion trigger: validation 対象、失敗条件、タグ taxonomy の管理境界を固定しないと保守運用に入れない段階になったら `Design Doc` 化する。

## Later

### Theme

`XDG_CONFIG_HOME/.pd/config.json` と project local な `.pd/config.json` で `kind` 語彙を決められるようにする。

- Outcome: 実装に埋め込まれた固定語彙だけでなく、利用者単位と repository 単位の設定から `kind` の許容値や運用語彙を定義できる。
- Why it matters: 現状の固定 `kind` 語彙だけでは project ごとの文書分類を表現しきれず、discovery のために必要な種別を repo 側で拡張できない。
- Horizon: Later
- Promotion trigger: global config と project local config のマージ規則、組み込み語彙との関係、validation と discovery 契約への反映範囲を固定しないと repo ごとの差分が実装に漏れ始める段階になったら `Design Doc` 化する。

### Theme

`pd` コマンド全般に human-readable 出力を追加する。

- Outcome: `--json` フラグなしで実行したとき、ターミナル向けの読みやすいテキスト形式で結果を表示できる。
- Why it matters: `--json` は machine / LLM 用途に最適化されているが、人間がターミナルで素早く確認したい場面では raw JSON は読みにくい。
- Horizon: Later
- Promotion trigger: 出力フォーマット仕様（カラム幅、省略ルール、カラー対応等）を固定する必要が出たら `Design Doc` 化する。

### Theme

`pd list` の大きい結果集合を段階取得できる公開契約を必要最小限で追加する。

- Outcome: 広い `--root` や多件数結果に対して、agent や CLI 利用者が一括取得前提ではなく複数回に分けて安全に取得できる方向性を持てる。
- Why it matters: 探索量そのものを抑える手段だけでは、結果集合が依然大きいケースで取得契約の負担が残る可能性がある。段階取得を roadmap に置いておくことで、`--depth` 導入後も不足が残るかを切り分けて判断できる。
- Horizon: Later
- Promotion trigger: `--depth` や既存 filter では agent-first な取得体験を十分に下げられず、結果の受け取り方自体を CLI 契約として固定しないと運用が不安定になる段階になったら `Design Doc` 化する。

### Theme

metadata curation を反復実行しやすい自動化経路へ接続する。

- Outcome: stale metadata の見直し、`related` 強化、タグ整備を CLI / skill / subagent から繰り返し実行できる。
- Why it matters: frontmatter は初回投入より継続的な手入れが価値の中心で、運用を自動化できないと corpus 品質がすぐ劣化する。
- Horizon: Later
- Promotion trigger: どの自動化経路でも Markdown frontmatter だけを書き戻す制約の下で、入出力契約を固定する必要が出たら `Design Doc` 化する。
