# 開発フロー

## フロー概要

```
PRD作成 → Issue作成 → 詳細設計 → TDD実装 → PRレビュー → マージ
```

## 各ステップ

### 1. PRD作成

- **担当Agent**: pdm
- **成果物**: `docs/prd/` 配下のPRDドキュメント
- PRDをPRとしてmainブランチにマージする
- `docs/prd/template.md` に沿って記述する

### 2. Issue作成

- **担当Agent**: pdm
- **成果物**: GitHubのIssue
- マージしたPRDをもとに実装Issueを作成する
- IssueにはPRDへの参照を含める

### 3. 詳細設計

- **担当Agent**: designer
- **成果物**: 設計ドキュメント・ADR・ドメインモデル
- Issueをもとに技術設計を行う
- 重要な意思決定は `docs/adr/template.md` に沿ってADRに残す
- ドメイン固有の概念は `docs/domain/template.md` に沿って定義する

### 4. TDD実装

- **担当Agent**: implementer
- **成果物**: テストコード・実装コード
- 設計ドキュメントのインターフェースに従って実装する
- Red-Green-Refactorサイクルを厳守する（テストを先に書く）

### 5. PRレビュー

- **担当Agent**: qa
- **成果物**: レビューコメント・承認
- テストカバレッジと要件との整合性を確認する

### 6. マージ

- レビュー承認後にmainブランチへマージする
