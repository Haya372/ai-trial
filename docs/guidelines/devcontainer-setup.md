# Dev Container セットアップガイド

## 概要

このプロジェクトはDev Containerを使った開発環境の隔離をサポートしています。
ホストマシンへのツールインストール不要で開発できます。

## 前提条件

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) または Docker Engine
- [VS Code](https://code.visualstudio.com/) + [Dev Containers 拡張機能](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)

## セットアップ手順

### 1. コンテナを起動する

VS Code でリポジトリを開き、コマンドパレット（`Cmd+Shift+P`）から以下を実行します：

```text
Dev Containers: Reopen in Container
```

初回起動時は以下が自動実行されます：

- Docker イメージのビルド（Ubuntu 24.04 + mise）
- PostgreSQL コンテナの起動
- `mise install` による全ツールのインストール
- `pnpm install` による Node.js 依存関係のインストール

### 2. 動作確認

コンテナ内で以下を実行してツールが正常にインストールされているか確認します：

```bash
# ツールバージョン確認
mise list

# Go ビルド確認
cd apps/backend && go build ./...

# フロントエンドビルド確認
mise exec -- pnpm build
```

## 開発中のコマンド実行

CLAUDE.md に記載の通り、pnpm コマンドは `mise exec -- pnpm <command>` で実行します。

```bash
# バックエンド起動
cd apps/backend && go run ./cmd/server

# フロントエンド起動
mise exec -- pnpm dev:web
```

## ポートフォワード

| ポート | 用途 |
|--------|------|
| 8080   | Backend API |
| 5173   | Frontend (Vite) |
| 5432   | PostgreSQL |

## Claude Code のサンドボックス設定

`.claude/settings.json` でサンドボックスが設定されています。
Dev Container 内では `bubblewrap` によるファイルシステム・ネットワーク分離が有効になります。

| 設定 | 値 | 説明 |
|---|---|---|
| `sandbox.enabled` | `true` | サンドボックス有効 |
| `sandbox.excludedCommands` | `["docker *", "gh *", "git *"]` | sandbox 非対応またはTLS問題があるコマンドを除外 |
| `sandbox.credentials.files` | `~/.aws`, `~/.ssh` | クレデンシャルファイルを読み取り禁止 |
| `sandbox.network.allowedDomains` | GitHub, npm, Go など | Bash コマンドが到達できるドメイン |

`enableWeakerNestedSandbox: true` は `postCreateCommand` によってコンテナ内の `.claude/settings.local.json` にのみ書き込まれます。ホスト直実行時には適用されません。

コンテナ内で Claude Code を起動することでコンテナとサンドボックスの二重隔離が得られます。

> **注意**: `docker-compose.yml` の `security_opt: apparmor=unconfined` は Ubuntu 24.04 で bubblewrap がユーザー名前空間を作成するために必要な設定です。AppArmor プロファイルを `bwrap` のみに限定して適用しており、コンテナ全体を無制限にするわけではありません。

```bash
# コンテナ内で Claude Code を起動
claude
```

## トラブルシューティング

### PostgreSQL に接続できない

`DB_HOST=postgres` が環境変数に設定されているか確認してください。
`devcontainer.json` の `remoteEnv` で自動設定されています。

### mise のツールが見つからない

```bash
mise install
```

を再実行してください。
