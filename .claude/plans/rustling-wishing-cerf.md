# rootの引数なし実行でヘルプ表示

## Context

現在`pd`を引数なしで実行するとkongがサブコマンド未指定エラーを出す。ユーザー体験として、引数なしで実行した場合はヘルプを表示してexit code 0で終了すべき。

## 実装方針

`main.go`で`kong.Parse`後に`k.Command()`を確認し、空文字列(サブコマンドなし)の場合は`kong.PrintUsage`でヘルプを表示して正常終了する。

また、ヘルプ出力を充実させるために`kong.Name("pd")`と`kong.Description(...)`を`kong.Parse`のオプションに追加する。

### 変更ファイル

1. **`main.go`** — `kong.Parse`後にサブコマンド未指定時のヘルプ表示ロジックを追加。`kong.Name`/`kong.Description`オプション追加。

```go
k := kong.Parse(new(cli.Root),
    kong.Name("pd"),
    kong.Description("Frontmatter discovery tool for project documentation."),
    kong.BindTo(ctx, (*context.Context)(nil)),
)

if k.Command() == "" {
    k.PrintUsage(false)
    os.Exit(0)
}

err := k.Run()
k.FatalIfErrorf(err)
```

2. **`internal/cli/cli_test.go`** — 引数なし実行のテストを追加。exit code 0かつstdoutにUsage情報が含まれることを検証。

## 検証方法

- `task check` で全体の lint + build + test を実行
- `go test -run TestCLI_NoArgs ./internal/cli/` で新規テストを実行
- 手動で `go run . ` を実行してヘルプが表示されることを確認
