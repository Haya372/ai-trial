# AI Trial

## プロジェクト概要

Claude Codeを活用した開発のテンプレートプロジェクト。

## 開発フロー

`docs/guidelines/development-flow.md` を参照。

## Agentの使い方

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

## Skillの使い方

各AgentはSkillを参照しながら作業する。Skillは `Skill` ツールで呼び出す。

## コマンド実行

- **pnpm**: corepack 経由だとシグネチャ検証エラーが出る。`mise exec -- pnpm <command>` で実行する

## 制約

- TDDを必ず守る（テストを先に書いてから実装する）
- Linterで守れるコーディング規約はAgentのコンテキストに含めない
- 技術スタック決定後は `.claude/settings.local.json` のHookにLinterコマンドを追加する
