# Workflow

## Flow

永続文書は `Roadmap`、`Design Doc`、`TODO.md` に絞る。`Plan Mode` は Codex / Claude Code の built-in planning 機能として使い、repo に恒久文書を増やさない。

流れは次の通り。

`Roadmap -> Design Doc -> TODO.md -> Plan Mode -> 実装`

更新原則は `append` ではなく `rewrite` とする。新しい論点が出たら追記せず、全体を削って書き直し、今の最小運用仕様だけを残す。

---

## Artifacts

### Roadmap

まだ `Design Doc` 化するには早いが、揮発させたくないアイデアを保持する。

- `Idea`
  何をやりたいか。まだ設計に落ちていない構想を短く書く。
- `Why it matters`
  それがなぜ必要か、何が前進するかを書く。
- `Promotion trigger`
  どの条件が揃ったら `Design Doc` 化するかを書く。

`Roadmap` には固定技術選定、実装順、細粒度タスク、`Reject if` を書かない。

### Design Doc

`TODO.md` に落とす前に固定すべき判断を持つ source of truth。

固定判断の canonical source は ADR とする。`Design Doc` は構造、制約、verification を持ち、固定判断は `ADR References` で参照する。判断が衝突した場合は ADR を優先する。

- `Goal`
  この設計で成立させたいことを書く。
- `Scope / Non-Scope`
  今回扱う範囲と、意図的に扱わない範囲を書く。
- `Architecture / Responsibility`
  どこに責務を置くか、どの境界で分けるかを書く。
- `ADR References`
  技術選定や採否判断として参照すべき ADR を列挙する。
- `Constraints`
  fail-closed、互換性、変更境界など、破ってはいけない制約を書く。
- `Verification Strategy`
  Theme の性質ごとに required verification coverage をどう決めるか、何で妥当性を確かめるかを書く。

`Design Doc` に残すのは、少なくとも次の入力になるものだけとする。

- `TODO Input`
  これがないと縦テーマを切れないもの。
- `Gate Input`
  これがないと `Reject if` や `Verification` を決められないもの。
- `Execution Input`
  これがないと `Plan Mode` で安全に実行単位へ切れないもの。

`Reference Only` は原則残さない。

特に技術選定は「現状説明」ではなく、「今回も前提として固定するか」「別案を採用しないか」という ADR で扱う。

次が未確定なら、まだ `TODO.md` に落とさない。

- 主要責務の置き場所
- 必要な ADR 判断
- verification taxonomy
- Theme 単位の required verification coverage

### TODO.md

未完了の縦テーマを管理する。TODO は実行命令書ではなく、テーマ管理のハブとする。

基本原則は、TODO を横分解ではなく縦テーマで書くこと。

- 良い例: ユーザーや外部から見て何が前進するかで切る
- 悪い例: parser、IR、renderer のように層や部品だけで切る

各 Theme は次を持つ。

- `Theme`
  何を前進させる縦テーマかを書く。
- `Outcome`
  終わると外から何ができるようになるかを書く。
- `Why now`
  なぜ今この Theme を進めるのかを書く。
- `Verification`
  Theme に必要な required verification coverage を書く。`static | unit | integration | system` を使い、`system` には e2e を含む。
- `Reject if`
  何なら今は採用不可かを書く。各項目は `[static]`、`[unit]`、`[integration]`、`[system]` の owner tag を持つ。
- `Why not split further?`
  なぜこの粒度で止めるのかを書く。

最小形は次。

```md
- [ ] Theme: ...
  - Outcome: ...
  - Why now: ...
  - Verification: static + integration
  - Reject if:
    - [static] ...
    - [integration] ...
    - [integration] ...
  - Why not split further?: ...
```

`Outcome`、`Verification`、`Reject if`、`Why not split further?` が弱い Theme は着手不可とする。

### Plan Mode

`TODO.md` の Theme を今回の実行単位へ一時的に分解する built-in ランタイム。`plan.md` は作らない。

`Plan Mode` では次だけを埋めればよい。

- 今回のスコープ
- 直近の実行順
- 先に失敗させる verification
- 今回の stop condition

Theme は、required verification coverage の全 gate に evidence があり、どの gate にも `reject-now` と `need-evidence` がなければ `Plan Mode` に進めてよい。`defer` は proceed を止めない。
required verification coverage の各 gate は必ず recorded disposition を出す。disposition がない gate は `need-evidence` とみなす。

---

## Design Doc -> TODO Rule

`Design Doc` から直接 task を作らず、まず縦テーマ候補を出す。候補は `Architecture / Responsibility` と `Verification Strategy` から列挙する。

TODO の単位は「終わると外から何が前進したか分かる縦テーマ」とする。層、部品、内部実装都合だけでは切らない。

`TODO.md` に落とせるのは次を満たす候補だけ。

- `Goal` とつながる
- `Outcome` が外から観測できる
- `Outcome` から required verification coverage を決められる
- `Reject if` が具体化できる
- required verification coverage の各 level に少なくとも1つの `Reject if` owner tag を割り当てられる
- 未確定の ADR 判断に依存しない
- `Why not split further?` に具体的に答えられる

複数候補がある場合の `Now` は、実装量の最小ではなく、設計の妥当性を最も早く検証できるものを選ぶ。

`Why not split further?` は分割停止条件として使う。これが書けないなら、Theme の粒度がまだ悪い。

required verification coverage は次で決める。

- 全 Theme: `static`
- 外から観測できる Outcome を持つ Theme: `static + integration`
- ユーザー価値や end-to-end flow を直接前進させる Theme: `static + integration + system`
- 純粋な補助 Theme に限り例外的に `static + unit`

`Reject if` は最低でも次の型に寄せる。

- `[integration]` 機能未達
- `[static|integration|system]` 禁止違反
- `[integration|system]` 回帰
- `[static|integration]` 境界逸脱
- `[static|unit|integration|system]` evidence 不足

---

## Review/Gate

review と gate は分ける。

- `Divergent Review`
  指摘候補を広く出す
- `Convergent Gate`
  `Reject if` と `Verification evidence` で裁定する

`Convergent Gate` の出力は次だけでよい。

- `reject-now`
- `need-evidence`
- `defer`

verification taxonomy は `static | unit | integration | system` に揃える。`system` は e2e を含む。Theme は required verification coverage の全 level を満たす必要がある。

- `static`
  静的解析、型、lint、禁止依存などの即時 gate。evidence は `executor/check identifier + case/suite/scenario identifier + pass/fail + replay handle`。
- `unit`
  局所責務、モジュール、関数、unit-level contract。evidence は `executor/check identifier + case/suite/scenario identifier + pass/fail + replay handle`。
- `integration`
  境界接続、API/CLI 契約、状態遷移、integration-level reject。evidence は `executor/check identifier + case/suite/scenario identifier + pass/fail + replay handle`。
- `system`
  主要シナリオ、ユーザー価値、stop-ship 条件、e2e-level reject。evidence は `executor/check identifier + case/suite/scenario identifier + pass/fail + replay handle`。

流れは次の通り。

1. `Divergent Review` が指摘候補を出す
2. 各粒度の `Convergent Gate` が、自分の owner tag を持つ `Reject if` と `Verification evidence` で裁く
3. `reject-now` と `need-evidence` だけが採用可否に影響する

`need-evidence` は required verification coverage に対する evidence contract のいずれかが欠けているときに使う。未実行またはラベルだけの identifier は evidence とみなさない。required gate では `defer` を使わない。

required verification coverage の全 gate に `reject-now` と `need-evidence` がなければ proceed してよい。

主 gate は次の順で使う。

1. static analysis
2. required verification coverage に含まれる unit / integration / system
3. gate owner は `static | unit | integration | system` のみとする

---

## Promotion Criteria

`WORKFLOW.md` を昇格済み文書として維持する条件は、次が即答できること。

- `Roadmap -> Design Doc -> TODO.md -> Plan Mode -> 実装` の流れ
- `ADR` と `Design Doc` が衝突したときの authority
- `Design Doc` に残すべき判断
- `Design Doc -> TODO` の切り方
- Theme が `Plan Mode` に進んでよい条件
- verification taxonomy
- `Reject if` の owner gate
- 各 verification level の evidence shape
- `Why not split further?` の役割
- `Divergent Review -> Convergent Gate` の流れ

次が残るなら rewrite 不十分とみなす。

- 同じ意味の節が複数ある
- 抽象語が多く、入力判定に落ちていない
- 実運用で使うテンプレートと昇格条件が見えない
- 追記の履歴が文書に残っている
