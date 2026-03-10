# TODO

- [ ] Theme: `pd list` の探索は `.gitignore` を尊重しつつ明示指定の `show` は維持する
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
