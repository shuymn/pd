# Done

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

- [x] Theme: 単一文書の metadata を見てから本文読みに進める
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

---

- [x] Theme: `pd list` の探索は `.gitignore` を尊重しつつ明示指定の `show` は維持する
  - Outcome: `pd list` は Git 管理下ではワークツリールート基準の ignore ルールを適用して `.md` 探索対象を絞り込む。ignore ソースは「リポジトリルートから探索ルートまでの各 `.gitignore`」「探索ルート配下のネストした `.gitignore`」「`.git/info/exclude`」に限定し、ignore されたディレクトリは剪定される。`pd show <path>` は ignore 状態に関係なく明示パス指定を優先して表示できる。Git 管理外ディレクトリでは ignore 判定は無効化され、現状どおり列挙される。CLI オプション追加なし、出力 JSON スキーマ変更なし。
  - Why now: 現状の `fs.WalkDir` 総当たりは `.gitignore` を無視して不要な `.md` まで探索しており、`pd list` の対象集合と走査量の両方にノイズを持ち込む。CLI や JSON 契約を変えずに探索境界だけを是正する最小テーマとして切り出せる。
  - Verification: static + integration
  - Reject if:
    - [static] ignore 判定が `list` 以外のコマンド経路に漏れて `show` の明示指定を阻害する
    - [static] matcher 初期化が `Scan` ごとに一度きりになっておらず、走査中に無駄な再構築を行う
    - [static] ignore パターンの評価基準がワークツリールート相対になっておらず、`--root` がサブディレクトリのときにルート側 `.gitignore` を正しく反映できない
    - [static] ignore ソース境界が曖昧で、`リポジトリルートから探索ルートまでの各 .gitignore`、`探索ルート配下のネストした .gitignore`、`.git/info/exclude` のどれを読むかが固定されていない
    - [integration] ルート `.gitignore` で無視された `.cache/` 配下の `.md` が `pd list` に出る
    - [integration] ignore されたディレクトリの剪定後に、無関係な通常ドキュメントまで列挙されなくなる
    - [integration] サブディレクトリの `.gitignore` がその配下以外にも誤って効く
    - [integration] negation パターンで再許可された `.md` を列挙できない
    - [integration] `--root docs/adr` のようなサブツリー走査で、リポジトリルート側 `.gitignore` が効かない
    - [integration] Git 管理外ディレクトリで ignore 機能が no-op にならず、従来列挙との差異が出る
    - [integration] ignore されたファイルへの `pd show <path>` が失敗する
    - [integration] ignore 対応のために CLI オプションや既存 JSON 出力の shape が変わる
  - Why not split further?: ignore matcher 構築、ディレクトリ剪定、`show` 非対象の境界は同一の探索仕様を構成している。個別に分けると「`list` だけが Git 互換に近い ignore を尊重する」という Outcome を統合的に検証できない。

- [x] Theme: `pd list` の diagnostic は通常非表示にし、`--verbose` でだけ観測できる
  - Outcome: `pd list` は invalid 文書が混在していても valid な discovery metadata だけを `stdout` に返し、通常実行では `stderr` を空に保って exit code `0` で成功する。`pd list --verbose` は同じ `stdout` に加えて invalid 文書ごとの diagnostic JSON を `stderr` に出す。`pd show` の failure は従来どおり `stderr` JSON と non-zero exit を維持する。
  - Why now: `.gitignore` 対応で `list` の探索対象を絞っても、invalid 文書の diagnostic が通常実行で大量に見えると LLM / automation 利用ではノイズが残る。探索境界の改善と独立に、CLI 可視性と終了コードの契約を固定する必要がある。
  - Verification: static + integration
  - Reject if:
    - [static] `Scan()` の `DiagnosticErrors` 契約まで崩して discovery 層の責務が変わる
    - [static] `--verbose` 判定が `show` に漏れて single-document failure の契約が変わる
    - [integration] invalid 文書が混在する `pd list --json` で `stderr` が空にならない
    - [integration] invalid 文書が混在する `pd list --json` が non-zero exit のままになる
    - [integration] invalid 文書しかない `pd list --json` が空配列を返さない
    - [integration] `pd list --verbose --json` で diagnostic JSONL が観測できない
    - [integration] `pd list --verbose --json` が diagnostic の有無で non-zero exit になる
    - [integration] `pd show` の invalid / not found failure が `stderr` JSON または non-zero exit を失う
  - Why not split further?: `list` の diagnostic 可視性、終了コード、`show` の failure 維持は同一の CLI error contract を構成している。分離すると「list は通常静かで、show は fatal を見せる」という利用時の挙動を一体で検証できない。
