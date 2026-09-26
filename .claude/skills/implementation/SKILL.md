---
name: implementation
description: TDD（Red-Green-Refactor）で実装するとき。
---

# 実装スキル

## 目的

設計ドキュメントのインターフェースを忠実に実装する。
Red-Green-Refactorサイクル自体は `superpowers:test-driven-development` スキルを使う。
このスキルはそれに加えて、このリポジトリ固有の前提・ルールを定義する。

---

## 前提

このスキルは以下が揃っていることを前提とする。まだ揃っていない場合は先に対応すること。

| 前提 | 確認方法 |
|---|---|
| 詳細設計ドキュメントが存在する | `docs/` 配下の設計ドキュメントを確認 |
| テストコードが先に書かれている（Red状態） | テストを実行し、意図した理由で失敗していることを確認 |

テストコードがまだない場合は、先に **test-design スキル** を使ってテストを書く。

---

## 手順

1. 詳細設計ドキュメントを読み、実装するインターフェース（関数シグネチャ・型・責務）を把握する
2. `superpowers:test-driven-development` スキルのRed-Green-Refactorサイクルに従って実装する
   - 設計ドキュメントのインターフェースを厳守する
   - 1サイクルを小さく保ち、完了したらコミットする（コミットメッセージは `commit` スキルの規約に沿う）
3. 次のテストケースを選び、1に戻る

---

## チェックリスト

実装完了前に以下を確認する。

- [ ] `superpowers:test-driven-development` の Verification Checklist を満たしているか
- [ ] 設計ドキュメントのインターフェースと実装が一致しているか
- [ ] コミットメッセージは `commit` スキルの規約に沿っているか
- [ ] コードコメントはWHY（理由）のみ書かれているか（WHATは書かない）

---

## HTTPハンドラーを実装するとき

HTTPハンドラーはユニットテスト（stub使用）に加え、**ルート統合テスト**も必須で実装する。

統合テストは既存の `*_route_integration_test.go` パターンに従い、testcontainersでDBを起動して実際のHTTPリクエスト/レスポンスを検証する。

最低限カバーすべきケース:

- 未認証時の 401 返却
- 正常系: 期待するレスポンス構造と値
- ユーザー分離: 他ユーザーのデータが混入しないこと

---

## フロントエンドのUI変更を実装するとき

画面の見た目や挙動に関わる変更を行った場合、テストのパス確認だけでなく、Playwright MCPを使って実際にブラウザ上で動作確認する。

1. `cd apps/backend && go run ./cmd/server` でバックエンドを起動する
2. `mise exec -- pnpm dev:web` で `apps/web` を起動する
3. Playwright MCPの `browser_navigate` で対象ページに遷移し、`browser_snapshot` 等で意図通り表示・動作しているか確認する

詳細は `docs/guidelines/playwright-mcp-guide.md` を参照。

---

## よくある落とし穴

| 落とし穴 | 対処 |
|---|---|
| ハンドラーのユニットテストだけで終わらせる | 統合テストも実装する（上記「HTTPハンドラーを実装するとき」参照） |
| 新規依存パッケージをキャレット等のレンジ指定で追加する | 追加先の `package.json` 内の既存の依存関係のバージョン指定方式（完全固定 or レンジ）に合わせる |
| READMEやコメントに「見ればわかること」を書く | コードコメントだけでなく、実装に伴い作成・更新するドキュメントにも `docs/guidelines/document-review-criteria.md` の観点を適用する |

---

## 参照ドキュメント

- `superpowers:test-driven-development` — Red-Green-Refactorサイクルの詳細
- `docs/guidelines/guidelines.md` — 設計原則
- `docs/guidelines/document-review-criteria.md` — ドキュメント作成時の観点（正確性・完全性・明瞭さ等）
- `docs/guidelines/development-flow.md` — 全体の開発フロー（PRD → Issue → 詳細設計 → TDD実装 → PRレビュー → マージ）
- `docs/guidelines/transaction-guidelines.md` — トランザクション管理（DBへの書き込みを伴うCommandを実装するとき）
- `docs/guidelines/playwright-mcp-guide.md` — Playwright MCPによるUI動作確認（フロントエンドのUI変更を実装するとき）
