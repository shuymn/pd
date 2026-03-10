# TODO

Source: [Frontmatter-Based Discovery Entry](docs/frontmatter-discovery-design.md)

- [x] Theme: metadata 一覧から kind ベースで候補文書を安定選別できる
  - Outcome: `pd list --json` が `docs/**/*.md` の discovery metadata を一覧で返し、`pd list --kind <kind> --json` が kind 絞り込み結果を返す。invalid document は stdout に混ざらず stderr JSON で報告される。
  - Why now: discovery reader、frontmatter decode、kind filtering、error contract の設計妥当性をまとめて検証できる最小テーマであり、show 系の前提基盤になる。
  - Verification: static + integration
  - Reject if:
    - [static] frontmatter schema を表す型と decode 経路が一致しない
    - [static] unknown field や unsupported field shape を silent accept する
    - [static] kind の語彙と validation 契約が一致しない
    - [integration] valid frontmatter を持つ複数文書から metadata 一覧を取得できない
    - [integration] kind で絞り込みできない
    - [integration] malformed frontmatter の文書が stdout に混入する
    - [integration] missing required field、unknown field、invalid kind、title 不在かつ H1 不在を reject できない
    - [integration] invalid document の error が stderr JSON で観測できない
    - [integration] CLI 出力契約が stdout success JSON / stderr error JSON として安定しない
  - Why not split further?: discovery reader と kind filtering は同一の read path 上にあり、reader だけでは外から観測可能な Outcome がない。error contract も list の出力契約と不可分であり、分離すると integration gate が成立しない。

- [ ] Theme: 単一文書の metadata を見てから本文読みに進める
  - Outcome: `pd show <path> --json` が対象文書の discovery metadata のみを返し、`pd show <path> --body` が本文まで含めて返す。対象不在や invalid frontmatter は stderr JSON で報告し non-zero exit で失敗する。
  - Why now: list 系で discovery reader が成立した後、body escalation の導線を検証する次の最小テーマ。metadata 確認後にだけ本文読みに進める二段階導線の設計妥当性を検証する。
  - Verification: static + integration
  - Reject if:
    - [static] show の decode 経路が list と同一の frontmatter schema 型を使わない
    - [integration] pd show --json が metadata のみを返さない
    - [integration] pd show --body が本文まで進めない
    - [integration] malformed frontmatter、missing required field、unknown field、invalid kind、title 不在かつ H1 不在、対象不在を reject できない
    - [integration] failure reason が stderr JSON で観測できない
    - [integration] single-document command が non-zero exit で失敗しない
  - Why not split further?: metadata 表示と body escalation は同一 path に対する段階的開示であり、片方だけでは「本文を読む前に metadata で判定する」という Outcome が成立しない。
