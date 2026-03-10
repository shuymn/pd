# /simplify レビュー結果と修正計画

## Context

`internal/discovery/gitignore.go` の新規追加と `discovery.go` への統合について、コード再利用・品質・効率の3観点でレビューした。

## レビュー結果サマリ

### コード再利用: 問題なし
- 既存ユーティリティとの重複なし。新コードはすべて gitignore 固有のロジック。

### コード品質

| # | 重要度 | 内容 | 対応 |
|---|--------|------|------|
| Q1 | Medium | `.git` がファイルの場合(worktree/submodule)に `info/exclude` やリポジトリルートの `.gitignore` が読めない | 対応不要 — 個人ドキュメントツールの用途では worktree/submodule シナリオは想定外。必要になった時点で対応。 |
| Q2 | Medium | ディレクトリツリーの二重走査(gitignore 収集 + 本体 Scan) | 後述 E1 と同一。修正する。 |
| Q3 | Low | symlink 未解決 | 対応不要 — docs ディレクトリで symlink は想定外。 |
| Q4 | Low | global gitignore (`core.excludesFile`) 未対応 | 対応不要 — 設計文書で ignore ソースを3種に限定済み。 |

### 効率

| # | 重要度 | 内容 | 対応 |
|---|--------|------|------|
| E1 | Medium | `loadDescendantPatterns` で事前に全サブディレクトリを walk → 本体 Scan で再度 walk の二重走査 | **修正する** |
| E2 | Medium | `loadDescendantPatterns` 内の matcher 再構築で `initialPatterns` を毎回コピー → O(n²) | **修正する** |
| E3 | Low | `repositoryIgnorer.Match` で毎回 slice 確保 | 対応不要 — 短命小スライスで GC 負荷は無視可能。 |

## 修正計画

### 修正1: 二重走査の解消 (E1/Q2)

`loadDescendantPatterns` を廃止し、メイン walk 中に `.gitignore` を遅延ロードする設計に変更する。

- `pathIgnorer` を stateful にし、ディレクトリ進入時に `.gitignore` を読み込む
- `repositoryIgnorer` に `enterDir(path string) error` 相当のロジックを追加、もしくは `Match` 内でディレクトリ初回アクセス時にパターンを遅延ロード
- `newPathIgnorer` では ancestor patterns + info/exclude のみロードし、descendant は walk 中に収集

**対象ファイル:**
- `internal/discovery/gitignore.go` — `loadDescendantPatterns` 廃止、`repositoryIgnorer` を stateful 化
- `internal/discovery/discovery.go` — walk 関数内で ignorer にディレクトリ通知

### 修正2: O(n²) パターンコピーの解消 (E2)

修正1で `loadDescendantPatterns` 自体が廃止されるため、この問題も同時に解消される。遅延ロード方式では matcher の再構築時に全パターンの単一スライスを grow するだけで済む。

## 検証

- `task check` (lint + build + test) がすべて通ること
- 既存テスト (`discovery_test.go` の gitignore 関連テスト含む) が変更なしで通ること
