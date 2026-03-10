# Convergent Gate: `pd list` の walk 範囲を `--root` 相対の `--depth` で制限できる

## Gate 総合裁定: **need-evidence**

proceed 不可。`reject-now` はないが、複数 gate で evidence が欠けている。

---

## Static Gate

| Reject-if | Disposition | 根拠 |
|---|---|---|
| `--depth` 責務が `list` 専用の探索境界から漏れ、共有 path 正規化や `show` の明示 path 解決に侵入する | **defer** | `Depth` は `Root` struct (global flag) にあるが、`show.go` は `root.Depth` を参照せず、`normalizePath` にも depth ロジックなし。depth 判定は `discovery.exceedsMaxDepth` / `relativeDepth` に集約。構造的には global だが機能的には list-only。厳密には `pd show --depth 0` がエラーなく受理される点が soft leakage だが、path 解決への侵入はない |
| 深さ判定の基準が discovery root 相対の単一ロジックに集約されず、呼び出し経路ごとに別実装が増える | **pass** | `exceedsMaxDepth` → `relativeDepth` → `splitRelativePath` の単一経路のみ。`handleDir` からのみ呼ばれる |
| required verification coverage の `static` gate に対する evidence が記録されない | **need-evidence** | executor/check identifier、pass/fail、replay handle の記録なし |

## Integration Gate

| Reject-if | Disposition | 根拠 |
|---|---|---|
| `--depth` の意味が discovery root 相対で固定されず、`--root` と組み合わせたときの深さ基準が曖昧 | **pass** | `TestCLI_List_Depth/explicit subtree root makes depth relative to subtree` が CLI 経由で検証 |
| 深さ超過ディレクトリを剪定できず、指定深さを超える walk を継続 | **pass** | `handleDir` が `fs.SkipDir` を返す。`TestScanner_Scan/depth prunes directories beyond max depth` で検証 |
| `pd list --depth 0 --json` が discovery root 直下の文書だけを安定して返さない | **pass** | `TestCLI_List_Depth/depth zero returns only root documents` が検証 |
| `pd list --depth 1 --json` が 1 階層下までの文書を返せない | **pass** | `TestCLI_List_Depth/depth one returns nested level` が検証 |
| `pd list --root <subtree> --depth 0 --json` で深さ基準が subtree 相対にならない | **pass** | 同上 subtree テスト |
| invalid な `--depth` 入力で非ゼロ終了にならない | **pass** | `negative depth exits non-zero` + `non integer depth exits non-zero` が検証 |
| depth 制限と kind filter を同時に指定したときに結果が壊れる | **pass** | `TestCLI_List_Depth/depth composes with kind filter` が検証 |
| `.gitignore` で除外される path の扱いが depth 対応で回帰する | **need-evidence** | unit テスト (`TestScanner_Scan/depth still respects gitignored directories`) のみ。CLI integration レベルの gitignore+depth テストが存在しない。Reject-if の owner tag は `[integration]` |
| `pd show --depth ... <path>` が既存どおりの明示取得を維持できない | **pass** | `TestCLI_Show_DepthDoesNotAffectExplicitPath` が検証 |
| required verification coverage の `integration` gate に対する evidence が記録されない | **need-evidence** | 記録なし |

## System Gate

| Reject-if | Disposition | 根拠 |
|---|---|---|
| agent-first な利用で広い `--root` に対して `--depth` を使った探索境界の抑制が効かない | **need-evidence** | system テスト未実装 |
| required verification coverage の `system` gate に対する evidence が記録されない | **need-evidence** | 記録なし |

---

## Divergent Review で検出された追加指摘 (Gate 外)

1. **README.md `pd show --depth 1` example が誤解を招く** — `show` は `--depth` を無視するが、README の usage example に含まれている (`README.md:76`)
2. **`Depth` が `int` (非ポインタ) で default:"3"** — CLI からは「無制限」を指定不可。設計判断として意図的かは要確認
3. **help テキストにデフォルト値 3 の記載なし** — kong が `--help` 出力で自動補完するかは未確認

---

## proceed に必要なアクション

1. **`task check` を実行** し、static gate (lint + build + test) の evidence を取得・記録する
2. **CLI integration テスト追加**: `.gitignore` + `--depth` の組み合わせテスト (`cli_test.go`)
3. **system テスト実施**: agent-first 利用シナリオ (広い `--root` + `--depth` で探索境界が効くことの e2e 検証)
4. **verification evidence 記録**: 全 gate (static / integration / system) の executor/check identifier、case/suite/scenario identifier、pass/fail、replay handle
5. **README.md 修正**: `pd show --depth 1` example を削除または注釈追加
