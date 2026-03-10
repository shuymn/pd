# Plan: `pd list` -- metadata 一覧から kind ベースで候補文書を安定選別

## Context

TODO.md の最初のテーマ。`pd` は設計完了・実装ゼロの状態（`main.go` がhello-worldのみ）。5つのADRで技術選定済み。このテーマで discovery reader、frontmatter decode、kind filtering、error contract の設計妥当性をまとめて検証する。

## Outcome

- `pd list --json` が `docs/**/*.md` の discovery metadata を一覧で返す
- `pd list --kind <kind> --json` が kind 絞り込み結果を返す
- invalid document は stdout に混ざらず stderr JSON で報告される

## Package Structure

```
main.go                              -- kong wiring, context.Background(), os.Exit
internal/
  cli/
    cli.go                           -- kong root struct
    list.go                          -- ListCmd + Run
  metadata/
    metadata.go                      -- Kind enum, Metadata struct, Result, Validate()
    metadata_test.go
  heading/
    heading.go                       -- ExtractH1() via goldmark
    heading_test.go
  frontmatter/
    frontmatter.go                   -- Extract() via adrg/frontmatter + yaml strict
    frontmatter_test.go
  diagnostic/
    diagnostic.go                    -- Diagnostic struct, Write()
    diagnostic_test.go
  discovery/
    discovery.go                     -- Scanner: walk + extract + validate + partition
    discovery_test.go
```

## Implementation Order (TDD)

### Step 1: `internal/metadata` -- 型とバリデーション
- `Kind` type + 8 enum values + `ParseKind(string) (Kind, error)`
- `Metadata` struct (yaml/json tags): `Kind`, `Description`, `Title`
- `Result` struct (json tags): `Path`, `Kind`, `Title`, `Description`
- `Validate(Metadata) error`: required kind/description, valid enum
- title 有無の判定は caller 責務（H1 fallback との組み合わせのため）

### Step 2: `internal/heading` -- H1 fallback
- `ExtractH1(body []byte) (string, bool)`
- goldmark で AST walk、最初の level-1 heading を返す
- ATX (`# Title`) と Setext (`===`) 両対応（goldmark がネイティブ処理）

### Step 3: `internal/frontmatter` -- frontmatter 抽出
- `Extract(r io.Reader, meta *metadata.Metadata) (body []byte, err error)`
- `frontmatter.MustParse` + custom Format with strict unmarshal for unknown field rejection
- `adrg/frontmatter` は内部で yaml.v2 を使用 → `yaml.UnmarshalStrict` を直接 import して unknown field rejection
- frontmatter なしは `frontmatter.ErrNotFound` で検出

### Step 4: `internal/diagnostic` -- エラー報告
- `Diagnostic{Command, Path, Reason string}` + JSON tags
- `Write(w io.Writer, d Diagnostic) error`

### Step 5: `internal/discovery` -- オーケストレーション
- `Scanner{Root string}` + `Scan(ctx, kind *Kind) ([]Result, []Diagnostic, error)`
- `fs.WalkDir` で `docs/**/*.md` を走査
- 各ファイル: frontmatter Extract → Validate → H1 fallback → Result or Diagnostic に振り分け
- 結果は path でソート（出力安定性）
- kind filter は Scan 内で適用

### Step 6: `internal/cli` + `main.go` -- CLI
- kong root struct with `List ListCmd`
- `ListCmd`: `--json` flag (required)、`--kind` optional filter
- `Run()`: Scanner 生成 → Scan → stdout に JSON array → stderr に diagnostic JSON lines
- `main.go`: `context.Background()`、kong parse、run、`os.Exit`

### Step 7: lint 設定更新
- `.golangci.yaml` exhaustruct include: `go-template` → `pd`
- kong struct を exhaustruct exclude に追加

## Dependencies (go.mod に追加)

- `github.com/alecthomas/kong` (ADR-009)
- `github.com/adrg/frontmatter` (ADR-008)
- `github.com/yuin/goldmark` (ADR-008)
- `gopkg.in/yaml.v2` (adrg/frontmatter の transitive dep、strict decode のため direct import)

## JSON Output Format

**stdout (成功):**
```json
[{"path":"docs/roadmap.md","kind":"roadmap","title":"Project Roadmap","description":"..."}]
```

**stderr (diagnostic、1行1JSON):**
```json
{"command":"list","path":"docs/broken.md","reason":"missing required field: kind"}
```

## Test Strategy

- **Unit (metadata, heading, diagnostic):** table-driven, `t.Parallel()`, I/O なし
- **Integration (frontmatter, discovery):** `t.TempDir()` に fixture .md ファイルを生成
- **CLI integration:** `exec.Command` でビルド済みバイナリを実行、stdout/stderr/exit code を検証
- fixture カテゴリ: valid (with title / with H1), invalid (no frontmatter / unknown field / unknown kind / missing kind / missing description / no title+no H1 / malformed YAML)
- 全テスト `-race -shuffle=on -count=10` 通過必須

## Technical Notes

- `frontmatter.MustParse` を使用（frontmatter なし = invalid）
- `yaml.UnmarshalStrict` で unknown field と duplicate key を reject
- `context.Context` を全 call path に threading（forbidigo 対応）
- `funlen` 80行制限 → `Scan` 内のロジックは `processFile` helper に分離
- docs root は CWD 相対で `docs/` をデフォルト

## Verification

```bash
task check          # lint + build + test (全 gate)
task test           # -race -shuffle=on -count=10
# CLI integration: テスト内で exec.Command("go", "run", ".") or TestMain でバイナリビルド
```
