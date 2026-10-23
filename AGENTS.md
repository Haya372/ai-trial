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

各工程の具体的な手順は `.agents/skills/` 配下に工程ごとのMarkdownファイル（`SKILL.md`）として置かれている。
Codex・Gemini CLIは `.agents/skills/` を自動で読み込む。Claude Code向けには `.claude/skills` を `.agents/skills` へのsymlinkにしている（Windowsでは `git config core.symlinks true` が必要）。
スキルを自動で読み込まないツールでも、該当する `SKILL.md` を直接読み、その手順に従うこと。
スキル本文中の「Claude Codeでは〜」という記述は、Claude Code以外のツールでは無視してよい。

| 工程 | 参照するファイル |
|---|---|
| Issue着手 | `.agents/skills/start-issue/SKILL.md` |
| Issue分割 | `.agents/skills/create-sub-issue/SKILL.md` |
| 要件定義・PRD作成 | `.agents/skills/requirements-definition/SKILL.md` |
| Issue起票 | `.agents/skills/create-issue/SKILL.md` |
| 仕様定義 | `.agents/skills/specification/SKILL.md` |
| 詳細設計 | `.agents/skills/detailed-design/SKILL.md` |
| テスト設計 | `.agents/skills/test-design/SKILL.md` |
| TDD実装 | `.agents/skills/implementation/SKILL.md` |
| リファクタリング | `.agents/skills/refactoring/SKILL.md` |
| コミット | `.agents/skills/commit/SKILL.md` |
| PR作成 | `.agents/skills/create-pr/SKILL.md` |
| スタックPR作成 | `.agents/skills/stack-pr/SKILL.md` |
| レビュー対応 | `.agents/skills/review-response/SKILL.md` |
| 振り返り | `.agents/skills/retrospective/SKILL.md` |
| ハーネス修正 | `.agents/skills/update-harness/SKILL.md` |

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

### コマンド実行時の注意

`mise exec -- pnpm <command>` 等のmise経由のコマンドは、mise が環境変数解決のために `.env.example` を読み込む。サンドボックスは `.env.*` を読み取り拒否対象にしているため、`.claude/settings.json` の `sandbox.filesystem.allowRead` で `.env.example` のみ読み取りを許可している。`allowUnsandboxedCommands: false` によりサンドボックスの無効化はできないため、mise経由のコマンドが別のサンドボックス制約で失敗した場合は、制約の内容をユーザーに報告して設定変更を依頼する。

### 制約

- 技術スタック決定後は `.claude/settings.local.json` のHookにLinterコマンドを追加する
