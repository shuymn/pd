# `--root` デフォルトを `"docs"` に変更し、git root 自動検出を導入

## Context

`discovery.Scanner.Scan` 内で `fs.WalkDir(os.DirFS(s.Root), "docs", walk)` と `"docs"` がハードコーディングされている。
git root を自動検出し、`--root` を git root からの相対パス（default: "docs"）とすることで、ハードコーディングを排除する。

## 設計

- CLI 層で git root を検出し、`filepath.Join(gitRoot, root.Root)` でスキャン対象の絶対パスを構築
- Scanner は解決済みのパスを受け取る（git 非依存を維持）
- Scanner 内では `filepath.Dir` / `filepath.Base` で FS ルートと walk パスを分離
- Result.Path は `"docs/a-doc.md"` 形式を維持（git root からの相対パス）

## 変更内容

### 1. `internal/cli/cli.go`
- `Root` フィールドの `default` を `"."` → `"docs"` に変更
- `help` テキストを「Directory to scan, relative to repository root.」に更新

### 2. `internal/cli/gitroot.go`（新規）
- `findGitRoot` 関数: `git rev-parse --show-toplevel` を実行して git root を取得
- git がインストールされていない場合やリポジトリ外で実行した場合はエラーを返す
- セキュリティ上の理由から自前のディレクトリトラバーサルではなく git 自身に委ねる

```go
func findGitRoot(ctx context.Context) (string, error) {
    out, err := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel").Output()
    if err != nil {
        // git root が見つからない場合は CWD にフォールバック
        return ".", nil
    }
    return strings.TrimSpace(string(out)), nil
}
```
```

### 3. `internal/cli/list.go`
- `Run` 内で `findGitRoot()` を呼び、`filepath.Join(gitRoot, root.Root)` を Scanner.Root に渡す

```go
func (cmd *ListCmd) Run(ctx context.Context, root *Root) error {
    gitRoot, err := findGitRoot()
    if err != nil {
        return err
    }

    s := discovery.Scanner{
        Root:   filepath.Join(gitRoot, root.Root),
        Logger: slog.New(diagnostic.NewHandler(os.Stderr)),
    }
    ...
}
```

### 4. `internal/discovery/discovery.go`

**`Scan` メソッド:**
```go
// Before
err := fs.WalkDir(os.DirFS(s.Root), "docs", walk)

// After
parent := filepath.Dir(s.Root)
dir := filepath.Base(s.Root)
err := fs.WalkDir(os.DirFS(parent), dir, walk)
```

**`handleFile` メソッド:**
```go
// Before
fullPath := filepath.Join(s.Root, path)

// After
fullPath := filepath.Join(filepath.Dir(s.Root), path)
```

### 5. `internal/discovery/discovery_test.go`
- `Scanner{Root: root}` → `Scanner{Root: filepath.Join(root, "docs")}` に変更
- ファイルは引き続き `root/docs/` 以下に配置
- パスアサーション (`"docs/a-doc.md"`) は変更不要
- 「no docs directory returns empty」テスト: Root に存在しないディレクトリを指定するよう変更

### 6. `internal/cli/cli_test.go`
- テスト用 temp dir に `.git` ディレクトリを作成（`os.MkdirAll(filepath.Join(root, ".git"), 0o755)`）
- または `--root` にフルパスを渡して git root 検出をバイパス...

→ CLI テストは `cmd.Dir = root`（temp dir）で実行。temp dir は git リポジトリではないため `findGitRoot` は CWD（`"."`）にフォールバック。`--root docs` により `root/docs` がスキャンされる。**既存テスト変更不要**。

### 7. `--root` は相対パスのみ

- `--root` は常に git root（または CWD フォールバック）からの相対パスとして解決
- 絶対パスはエラー
- `..` を含むパス（ディレクトリトラバーサル）はエラー
- `filepath.Clean` 後に上記チェックを行う（`foo/../bar` → `bar` のように正規化される前に `..` をチェック）

```go
// list.go の Run 内（または validateRoot ヘルパー）
func validateRoot(root string) error {
    if filepath.IsAbs(root) {
        return fmt.Errorf("--root must be a relative path, got %q", root)
    }
    // Clean前に".."を含むかチェック（Clean後だと foo/../bar が bar になり検出不可能なケースがある）
    // Clean後のパスでも".."で始まるパスを拒否
    cleaned := filepath.Clean(root)
    if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
        return fmt.Errorf("--root must not traverse above the base directory, got %q", root)
    }
    return nil
}
```

### 8. テスト追加

**`internal/cli/cli_test.go`:**
- `--root /absolute/path` → エラー
- `--root ../outside` → エラー
- `--root foo/../../outside` → エラー

## 検証

```sh
task check
```
