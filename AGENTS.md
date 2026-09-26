# AI Trial

Claude Code・Codex・Gemini CLIなど、AGENTS.md対応のAIコーディングエージェント共通の規約。

## プロジェクト概要

Claude Codeを活用した開発のテンプレートプロジェクト。

## 開発フロー

`docs/guidelines/development-flow.md` を参照。

PRD作成 → Issue作成 → 仕様定義 → 詳細設計 → TDD実装 → PRレビュー → マージ、の順で進める。

## Agentの使い方（Claude Code以外のツール）

Claude Codeでは工程ごとに専用のSubagent（pdm/designer/implementer/qa等）へ切り替えながら進める。
Subagentの仕組みを持たないツールでは、1つのセッション内で開発フローの工程を順番に進めればよい。

各工程の具体的な手順は `.claude/skills/` 配下に工程ごとのMarkdownファイル（`SKILL.md`）として置かれている。
Claude CodeのSkillツールを持たないツールでも、該当する `SKILL.md` を直接読み、その手順に従うこと。

| 工程 | 参照するファイル |
|---|---|
| Issue着手 | `.claude/skills/start-issue/SKILL.md` |
| 要件定義・PRD作成 | `.claude/skills/requirements-definition/SKILL.md` |
| Issue起票 | `.claude/skills/create-issue/SKILL.md` |
| 仕様定義 | `.claude/skills/specification/SKILL.md` |
| 詳細設計 | `.claude/skills/detailed-design/SKILL.md` |
| テスト設計 | `.claude/skills/test-design/SKILL.md` |
| TDD実装 | `.claude/skills/implementation/SKILL.md` |
| リファクタリング | `.claude/skills/refactoring/SKILL.md` |
| コミット | `.claude/skills/commit/SKILL.md` |
| PR作成 | `.claude/skills/create-pr/SKILL.md` |
| スタックPR作成 | `.claude/skills/stack-pr/SKILL.md` |
| レビュー対応 | `.claude/skills/review-response/SKILL.md` |
| 振り返り | `.claude/skills/retrospective/SKILL.md` |
| ハーネス修正 | `.claude/skills/fix-harness/SKILL.md` |

## コマンド実行

- **pnpm**: corepack 経由だとシグネチャ検証エラーが出る。`mise exec -- pnpm <command>` で実行する

## 制約

- TDDを必ず守る（テストを先に書いてから実装する）
- Linterで守れるコーディング規約はAgentのコンテキストに含めない
- 新規ドキュメント（ガイドライン・review-criteria等）を作成する前に、`docs/guidelines/` 配下の既存ドキュメントを確認し、内容の重複を避ける

## Claude Code固有

以下はClaude Code専用の設定。他ツールでは無視してよい。

### Agentの使い方

各Agentは責務が絞られている。工程ごとにAgentを切り替えながら進める。
1つのAgentに長期間作業させず、コンテキストを節約する。

| Agent | 役割 | 呼び出すタイミング |
|---|---|---|
| pdm | 要件定義・PRD作成・仕様定義 | 開発の起点、機能要件の整理、仕様書作成 |
| designer | 詳細設計・ADR作成 | 仕様定義後に技術設計するとき |
| implementer | TDD実装 | 設計完了後に実装するとき |
| advisor | 技術アドバイス・設計レビュー | 判断に迷ったとき |
| qa | テスト・PRレビュー | マージ前の品質確認 |
| persona | ユーザー視点のレビュー | 要件の妥当性を確認するとき |
| explorer | コード調査・解析（Haiku使用） | 大規模調査でコンテキストを節約したいとき |

### Skillの使い方

各AgentはSkillを参照しながら作業する。Skillは `Skill` ツールで呼び出す。

### 制約

- 技術スタック決定後は `.claude/settings.local.json` のHookにLinterコマンドを追加する
