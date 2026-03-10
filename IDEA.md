# `pd` / Frontmatter / Progressive Disclosure 設計メモ

## 1. 目的

リポジトリ内の `docs/**/*.md` を、LLM が `ls` / `cat | head` ベースの偶然的な探索ではなく、より安定した **Progressive Disclosure** で読めるようにする。

そのために各 Markdown に discovery 用の frontmatter を持たせ、`pd` コマンドを通じて以下を実現する。

* まず軽い metadata だけを見て「読むべき文書か」を判断する
* 必要な文書だけ本文を読む
* 文書群の metadata を後から独立した保守作業として更新できる
* それでも **SSoT は常に Markdown 本体** のまま保つ

---

## 2. 結論

### 2.1 SSoT

**Markdown ファイルが唯一の SSoT である。**

* 本文も metadata も Markdown ファイルに存在する
* 文書固有 metadata は frontmatter に置く
* `pd` は Markdown を読む/更新する補助ツールであり、第二の知識ベースではない
* 外部の手管理 index や metadata store は持たない

### 2.2 `pd` の位置づけ

`pd` は次の 3 つを担う。

* **reader**: frontmatter を安定して読む
* **writer**: frontmatter を構造化更新する
* **curation helper**: discovery metadata の保守運用を支援する

つまり `pd` は単なる検索ツールでも、RAG 基盤でも、MCP サーバでもない。

**Markdown repo のための deterministic helper / metadata maintenance CLI** として設計する。

### 2.3 frontmatter の方針

frontmatter は本文の要約ではなく、**LLM が読むべきかどうか判断するための metadata** に寄せる。

そのため、`summary` のような本文縮約は基本的に採用しない。

代わりに、以下のような「発火条件」「読むべき条件」に寄せた情報を持つ。

* `description`
* `read_when`
* `not_for`
* `tags`
* `related`

---

## 3. 背景

### 3.1 なぜ `ls` / `cat | head` では弱いのか

`ls` / `cat | head` は人間向けの雑な探索には十分だが、LLM にとっては以下の問題がある。

* ファイル名や冒頭段落の偶然性に強く引っ張られる
* 文書種別や正本性が分からない
* deprecated な文書やノイズを拾いやすい
* 本文を読む前の discovery 信号が弱い

### 3.2 Progressive Disclosure 的に何をやりたいか

理想の探索手順は以下。

1. 文書一覧の metadata を見る
2. いま読むべき文書かを判断する
3. 候補を絞る
4. 必要な文書だけ本文を開く
5. 必要なら related を辿る

この手順を安定化するための入口が `pd` である。

---

## 4. 設計原則

### 4.1 Markdown SSoT 原則

* truth は常に Markdown にある
* frontmatter も本文と同じく Markdown に属する
* `pd` は Markdown を直接読む/更新する
* 手管理の外部 index は持たない
* 生成キャッシュは持ってよいが、編集対象にしない

### 4.2 Metadata は discovery 専用に寄せる

frontmatter は本文の別表現ではなく、**routing / discovery 用 metadata** として使う。

#### 採用しない方向

* 長い要約
* 本文の縮約コピー
* 本文と同じ情報の重複保持

#### 採用する方向

* いつ読むべきか
* 何には使わないか
* どのトピックに属するか
* 関連文書は何か
* その metadata をいつ見直したか

### 4.3 `pd` は補助に徹する

`pd` は本文の代替ではない。

* 本文の意味や設計判断は本文側が持つ
* `pd` は frontmatter の read / write / validate / curation を担う
* 本文の構造を変えても、`pd` は frontmatter を読むだけなので柔軟性を損なわない

### 4.4 本文執筆と metadata キュレーションを分離する

すべての metadata を文書執筆者が最初から丁寧に書く必要はない。

特に以下は、後から独立したエージェントや review タイミングで更新してよい。

* `description`
* `read_when`
* `not_for`
* `tags`
* `related`
* `metadata_reviewed_at`

この分離により、本文執筆と discovery 品質改善を別工程として扱える。

---

## 5. frontmatter 設計

## 5.1 フィールド分類

### A. core fields

文書の基本的な性質を表す。著者またはメンテナが持つ。

* `title`
* `kind`
* `status`
* `canonical`
* `supersedes`（必要なら）

### B. discovery fields

LLM が「読むべきか」を判断するための metadata。後段の curator が更新してよい。

* `description`
* `read_when`
* `not_for`
* `tags`
* `related`
* `metadata_reviewed_at`

### C. body

実際の設計判断・仕様・手順・背景などの本体。

---

## 5.2 `summary` を採用しない理由

`summary` は一見便利だが、今回の方針では不採用寄りとする。

理由:

* 本文の縮約になりやすい
* frontmatter が本文の重複になる
* 更新漏れすると壊れやすい
* discovery ではなく content duplication になる
* LLM が「読むべき条件」ではなく「内容っぽい文」に引っ張られる

もし人間向けの要約が欲しい場合は、frontmatter ではなく本文冒頭や `## Purpose` など本文側で持つ。

---

## 5.3 推奨 frontmatter 例

```yaml
---
title: Auth Session Architecture
kind: design-doc
status: active
canonical: true

description: Use for authentication/session design changes and session inconsistency investigations.
read_when:
  - ログインまわりの責務分担を変更するとき
  - セッション不整合を調査するとき
not_for:
  - UI文言だけの変更

tags:
  - auth
  - session
  - token

related:
  - docs/runbooks/session-debug.md
  - docs/adr/adr-0012-session-store.md

metadata_reviewed_at: 2026-03-09
---
```

---

## 6. `pd` の責務

## 6.1 読み取り

`pd` は Markdown を走査し、frontmatter を安定して取り出す。

目的:

* `ls` / `head` の代替 discovery 入口
* 機械可読な一覧取得
* `status` / `kind` / `canonical` / `tags` などによる絞り込み
* 本文を開く前の判断材料提供

## 6.2 更新

`pd` は frontmatter を構造化更新する。

目的:

* LLM が YAML を全文再生成しないようにする
* 更新と validation を一体化する
* field 順序や style のブレを抑える
* 部分更新を安全に行う

重要なのは、`pd` が外部 metadata を更新するのではなく、**Markdown 本体の frontmatter を直接編集する**こと。

## 6.3 検証

`pd` は schema / style / 整合性を検証する。

例:

* 必須 field の欠落
* `canonical` の重複
* `related` のリンク切れ
* `status` と `supersedes` の整合性
* タグ語彙の逸脱

## 6.4 キュレーション支援

`pd` の重要な責務はここにある。

`pd` は frontmatter の CRUD だけでなく、**discovery metadata の保守運用**を支援する。

例:

* metadata の review queue を出す
* 古い文書から順に見直せるようにする
* 既存タグ一覧を出して語彙合わせを支援する
* `related` 候補を提案する
* curated fields だけを後段 agent に更新させる

---

## 7. `pd` コマンド面（最小案）

## 7.1 読み取り系

```bash
pd list --json
pd list --kind design-doc --status active --json
pd show docs/design/auth-session.md --json
pd show docs/design/auth-session.md --body
pd related docs/design/auth-session.md --json
```

### 意図

* `list`: metadata 一覧
* `show`: 1 文書の frontmatter を確認
* `show --body`: 必要時だけ本文へ進む
* `related`: 関連文書を追う

---

## 7.2 更新系

```bash
pd get docs/design/auth-session.md
pd set docs/design/auth-session.md status active
pd set docs/design/auth-session.md canonical true
pd add docs/design/auth-session.md tags auth
pd add docs/design/auth-session.md related docs/runbooks/session-debug.md
pd remove docs/design/auth-session.md tags experimental
pd apply docs/design/auth-session.md --from patch.json
pd init docs/design/new-doc.md --kind design-doc
```

### 意図

* LLM が frontmatter 全文を手書きしない
* YAML を自由記述させず、構造化 patch に寄せる
* 更新時に normalize / validate を内包する

---

## 7.3 キュレーション系

```bash
pd review queue --sort metadata_reviewed_at --older-than 90d
pd curate docs/design/auth-session.md --fields description,read_when,not_for,tags,related
pd tags list --count
pd tags suggest docs/design/auth-session.md
pd tags rename authentication auth
pd tags lint
pd related suggest docs/design/auth-session.md
pd validate docs/design/auth-session.md
pd validate --all
```

### 意図

* 後段の review / curation 作業を支援する
* タグ語彙の cardinality を制御する
* 文書間リンクを後から育てる
* stale な discovery metadata を順に更新できるようにする

---

## 8. review / curation 運用

## 8.1 metadata review 日付

frontmatter には本文全体のレビュー日ではなく、**discovery metadata を見直した日** を持たせる。

推奨 field:

* `metadata_reviewed_at`

これにより、`pd` は「古い順に見直す」運用を支援できる。

### 例

* 90 日以上見直されていない docs を列挙
* `status: active` かつ `metadata_reviewed_at` が古いものを優先
* curator agent が順番に更新する

---

## 8.2 本文執筆と review を分ける

推奨フロー:

1. 著者が本文と core fields を書く
2. 必要最小限の frontmatter でコミットする
3. 別タイミングで curator agent が `pd curate` を使う
4. `description / read_when / not_for / tags / related / metadata_reviewed_at` を更新する
5. `pd validate --all` または CI で整合性確認

このフローにより、著者の負担を増やしすぎずに discovery 品質を上げられる。

---

## 9. タグ設計

## 9.1 `keywords` ではなく `tags`

`keywords` という名前だと free-form な単語が増えやすい。

今回必要なのは曖昧検索用の語ではなく、**低カーディナリティな分類語彙** なので `tags` を採用する。

### ねらい

* corpus 全体で語彙を揃える
* 似ているが違う単語の乱立を防ぐ
* 既存 docs と雰囲気を合わせる
* filtering / grouping を安定させる

---

## 9.2 タグ運用方針

* なるべく低カーディナリティ
* 再利用される語に寄せる
* singular noun を基本にする
* 近いが違う語を増やしすぎない
* review 時に既存タグ一覧を見て揃える

### 例

推奨:

* `auth`
* `session`
* `token`
* `payments`
* `cache`

避けたい例:

* `authentication`
* `sessions`
* `access-token`

---

## 9.3 タグ語彙の SSoT

タグ語彙の指針も Markdown に置く。

例:

`docs/meta/tag-taxonomy.md`

ここに書くもの:

* 推奨タグ一覧
* 避けるタグ
* 置換ルール
* 運用ルール

これにより、タグ語彙も Markdown SSoT の原則を保てる。

---

## 10. `related` の扱い

`related` は著者が最初に書いてもよいが、**後段の curator が強化する前提**で考える。

理由:

* 新しい文書が後から増える
* 同テーマの別種別 docs は後日できることが多い
* 著者は自分の文書中心で見ている
* 横断的な見渡しは curator 側の方が強い

したがって `related` は review タイミングで更新していく対象として扱う。

---

## 11. README / AGENTS.md の役割

README / AGENTS.md に書くべきなのは **規約** であり、文書データそのものではない。

### README / AGENTS.md に置くもの

* docs をどう読むか
* `pd` をまず使うこと
* canonical / active を優先すること
* frontmatter の意味
* 本文を広く読む前に metadata を確認すること

### frontmatter に置くもの

* 文書固有 metadata
* discovery 用 metadata

### 置かないもの

* 手管理の repo-wide 文書 index
* frontmatter と重複する metadata の別管理

---

## 12. 将来の自動化

`pd` は単独 CLI として使えるが、将来的には以下と組み合わせられる。

* Skill
* subagent
* prompt 出力補助

### 適した用途

#### Skill / subagent

* repo を読みながら `pd` を叩く反復的なキュレーション
* `description` / `tags` / `related` の再整備
* stale docs の review queue 処理

#### prompt 出力補助

* モデル非依存の review prompt を生成したいとき
* 人手 review / 他エージェント review に回したいとき

ただし、どの自動化経路を使っても、書き戻し先は常に Markdown frontmatter とする。

---

## 13. 非目標

この設計で最初からやらないこと:

* 外部の手管理 metadata DB
* 手管理の `index.yml`
* vector search / RAG / reranker 前提の大規模検索基盤
* 本文の要約を frontmatter に重複保持すること
* LLM に YAML 全文を自由記述させること
* 本文内容そのものを `pd` が解釈して正誤判断すること

---

## 14. 実装指針

## 14.1 基本方針

* `.md` が SSoT
* `pd` は frontmatter を parse / patch / validate する
* 機械可読出力は `--json` を第一級にする
* 書き戻しは Markdown 本体に対してのみ行う
* キャッシュを持つなら生成物として扱う

## 14.2 更新の基本姿勢

* free-form rewrite より structured mutation
* `set` / `add` / `remove` / `apply` を中心にする
* 更新と validation を一体化する
* field 順序や重複を normalize する
* 本文は触らない

## 14.3 安全策

* unknown field を勝手に消さない
* schema 違反なら書き戻さない
* `--dry-run` や diff 表示を持つ
* `pd validate --all` を CI に載せられるようにする

---

## 15. 現時点の最終像

### 一文で言うと

**`pd` は、Markdown を SSoT としたまま、frontmatter ベースの Progressive Disclosure と discovery metadata の保守運用を支える CLI である。**

### そのための原則

* 真実は Markdown にのみ置く
* frontmatter は本文要約ではなく discovery metadata に寄せる
* `summary` は基本採用しない
* `pd` は read / write / validate / curate を担う
* metadata review は本文執筆と分離できる
* タグは低カーディナリティな制御語彙として扱う
* `related` や `description` は後から育ててよい

---

## 16. 次に詰めるべきこと

1. frontmatter schema の最小必須項目を決める
2. `pd` の v1 コマンドセットを固定する
3. `pd validate` のルールを決める
4. `tags` の運用ルールと taxonomy Markdown を決める
5. `pd curate` の入出力形式を決める
6. AGENTS.md に docs discovery protocol を書く
