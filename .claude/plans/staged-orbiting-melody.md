# /simplify レビュー結果と修正計画

## Context

`internal/discovery/gitignore.go` の新規追加と `discovery.go` への統合について、コード再利用・品質・効率の3観点でレビューした。

## レビュー結果サマリ

### コード再利用: 問題なし

- `parseIgnoreFile` は go-git の `readIgnoreFile` と類似するが、利用には `billy.Filesystem` 依存が必要で、~25行の節約に見合わない
- `findGitWorkTreeRoot` は go-git の unexported 関数と類似するが、`PlainOpen` は重すぎるため独自実装は妥当
- `splitRelativePath`, `ancestorDomains` は既存コードに重複なし

### コード品質

| # | 重要度 | 内容 | 対応 |
|---|--------|------|------|
| Q1 | Medium | patterns/matcher の二重状態管理 — matcher 再構築忘れで不整合リスク | 対応不要 — `EnterDir` のみが変更し即座に再構築するため実害なし。構造体フィールド2つで十分シンプル |
| Q2 | Low | `parseIgnoreFile` が `\#` エスケープ未対応 | 対応不要 — ドキュメントツール用途で `\#` パターンは想定外 |
| Q3 | Low | global gitignore (`core.excludesFile`) 未対応 | 対応不要 — 設計文書で ignore ソースを限定済み |

### 効率

| # | 重要度 | 内容 | 対応 |
|---|--------|------|------|
| E1 | Medium | patterns が単調増加 — 退出済みディレクトリのパターンも `Match` で毎回走査 | 対応不要 — go-git の domain スコーピングで正確性は担保。ドキュメントリポジトリの規模では影響軽微 |
| E2 | Low-Medium | `Match` で毎回 `[]string` スライス確保 | 対応不要 — `WalkDir` はシングルスレッドで GC 負荷は軽微。プロファイル結果なしに最適化は早計 |
| E3 | Low | `splitRelativePath` 内の `filepath.Clean` が walk パスでは冗長 | 対応不要 — コスト無視可能 |

## 結論

**修正すべき問題は見つからなかった。** コードは既にクリーンな状態。

- コード再利用: 既存コードとの重複なし、go-git の代替は依存コスト対効果で不採用が妥当
- コード品質: patterns の単調増加は設計上の選択で、domain スコーピングにより正確性は担保される
- 効率: ドキュメントリポジトリの規模ではすべて影響軽微。プロファイリングで問題が確認された場合に対応
