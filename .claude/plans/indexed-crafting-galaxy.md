# diagnostic パッケージを slog ベースに置き換え

## Context

ADR-003 で `log/slog` 採用が決定済みだが未実装。現在 `internal/diagnostic` パッケージが自前の `Diagnostic` 構造体と `Write` 関数で JSONL 出力を行っている。これを slog ベースに置き換え、ADR-007 の stderr 契約（`{"path":"...","reason":"..."}` 形式）はカスタム Handler で維持する。

## 方針

- `internal/diagnostic` パッケージを「カスタム slog.Handler の提供元」に変更
- `discovery.Scanner` に `*slog.Logger` フィールドを追加し、diagnostic を slog 経由で出力
- `Scan()` の戻り値から `[]diagnostic.Diagnostic` を除去し `([]metadata.Result, error)` に変更
- CLI 層で Handler を生成し Logger を Scanner に渡す

## 変更ファイル

### 1. `internal/diagnostic/diagnostic.go` — 全面書き換え

`Diagnostic` 構造体と `Write` 関数を削除。代わりにカスタム `slog.Handler` を実装。

```go
package diagnostic

type Handler struct {
    mu  sync.Mutex
    w   io.Writer
    enc *json.Encoder
}

func NewHandler(w io.Writer) *Handler

// slog.Handler interface
func (h *Handler) Enabled(context.Context, slog.Level) bool   // WARN 以上
func (h *Handler) Handle(ctx context.Context, r slog.Record) error
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler
func (h *Handler) WithGroup(name string) slog.Handler
```

`Handle` は Record の Attrs から `path` と `reason` を抽出し、`{"path":"...","reason":"..."}` をJSONLとして書き出す。`WithAttrs`/`WithGroup` は使わないがインターフェース充足のため実装（no-op または自身を返す）。

### 2. `internal/diagnostic/diagnostic_test.go` — 全面書き換え

カスタム Handler のテスト:
- `slog.New(handler).Warn("msg", "path", "docs/x.md", "reason", "error text")` の出力が `{"path":"...","reason":"..."}` 形式であること
- 複数回の Handle 呼び出しで JSONL（1行1JSON）になること
- INFO レベルは出力されないこと（Enabled で弾く）

### 3. `internal/discovery/discovery.go` — シグネチャ変更 + slog 利用

- `Scanner` に `Logger *slog.Logger` フィールドを追加
- `Scan()` を `([]metadata.Result, error)` に変更（`[]diagnostic.Diagnostic` を除去）
- `handleFile` / `processFile` で diagnostic 生成箇所を `logger.WarnContext(ctx, "...", "path", relPath, "reason", reason)` に変更
- `processFile` は `(*metadata.Result, error)` を返す。invalid doc の場合は `nil, nil` を返し（呼び出し元で logger 出力済み）、fatal error のみ `error` を返す
  - 具体的には: `processFile` は `(*metadata.Result, string, error)` を返す（string が diagnostic reason、空なら diagnostic なし）。`handleFile` で reason が非空なら logger で出力
- Logger が nil の場合は `slog.New(slog.DiscardHandler)` をフォールバックとして使用（`Scan` 冒頭で）

### 4. `internal/discovery/discovery_test.go` — 戻り値変更対応

- `Scan()` の戻り値を2つに変更（`results, err`）
- diagnostic 数の検証は、テスト用バッファに書く Handler 経由の slog.Logger を Scanner に渡して行う
  - `var buf bytes.Buffer` + `diagnostic.NewHandler(&buf)` → `slog.New(handler)` を `Scanner.Logger` にセット
  - バッファの行数で diagnostic 数を検証
  - 各行を JSON パースして path/reason の非空を検証

### 5. `internal/cli/list.go` — Logger セットアップ、ループ削除

- `diagnostic.NewHandler(os.Stderr)` で Handler 生成
- `slog.New(handler)` を `Scanner.Logger` にセット
- diagnostic イテレーションループを削除（slog が Scan 中に直接 stderr へ出力するため）
- `Scan` の戻り値を2つに変更

### 6. `internal/cli/cli_test.go` — 変更なし

統合テストはバイナリ経由で実行しており、stderr の JSONL フォーマットは維持されるため変更不要。

## ドキュメント影響分析

今回の変更は外部契約（stderr JSONL with path+reason）を維持したまま内部実装を差し替えるため、以下はすべて変更不要:

- `TODO.md` — 外部契約（"stderr JSON で報告"）のみ記述。実装手段に言及なし
- `docs/frontmatter-discovery-design.md` — Error contract セクションは外部契約のみ定義
- `docs/roadmap.md` — diagnostic の実装詳細への言及なし
- `docs/adr/003-slog-for-logging.md` — この変更はまさに ADR-003 の実装
- `docs/adr/007-frontmatter-invalid-discovery-behavior.md` — 契約はカスタム Handler で維持

## 変更しないファイル

- `main.go` — 変更不要
- `internal/cli/cli.go` — 変更不要

## リスク

- **出力タイミングの変化**: 現在は全 diagnostic を蓄積してから一括出力。slog 版では Scan 中にリアルタイム出力。stdout は Scan 完了後に書かれるため、stdout/stderr の混在問題はない。
- **Handle のエラー**: slog は `Handle` のエラーを黙殺する。`json.Encode` が2つの string で失敗する可能性は極めて低いため許容。

## 検証

1. `go test ./internal/diagnostic/...` — Handler 単体テスト
2. `go test ./internal/discovery/...` — Scanner テスト
3. `task check` — lint + build + 全テスト（統合テスト含む）
