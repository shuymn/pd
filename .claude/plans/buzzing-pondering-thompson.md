# Simplify: `--depth` オプション追加コードのレビュー

## Context

`pd list` に `--depth` オプションを追加する変更に対して、コード再利用・品質・効率の3観点でレビューを実施。

## 指摘事項

### 1. デッドコード: `isDir` パラメータ (品質)

- **場所:** `internal/discovery/discovery.go:200,208`
- **問題:** `exceedsMaxDepth` と `relativeDepth` は常に `isDir=true` で呼ばれる（`handleDir` からのみ呼出）。`isDir` パラメータと `false` 分岐は到達不能コード。
- **修正:** 両関数から `isDir` パラメータを削除し、常にディレクトリ前提のロジックにする。

```go
func exceedsMaxDepth(path string, maxDepth *int) bool {
	if maxDepth == nil || path == "." {
		return false
	}
	return relativeDepth(path) > *maxDepth
}

func relativeDepth(path string) int {
	return len(splitRelativePath(path))
}
```

### 2. README の誤解を招く例 (品質)

- **場所:** `README.md:76`
- **問題:** `pd show --depth 1 docs/adr/001.md` という例があるが、設計上 `show` は `--depth` を無視する。ユーザーに誤った期待を与える。
- **修正:** この行を削除する。

## 修正対象ファイル

- `internal/discovery/discovery.go` — `isDir` パラメータ削除
- `README.md` — 誤解を招く `show --depth` 例の削除

## 検証

- `task check` で lint + build + test が全て通ることを確認
