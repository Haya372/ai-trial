# taskctl

タスクごとに隔離コンテナを起動・停止するラッパー。方式は [ADR-035](../../docs/adr/ADR-035-orchestrator-subagent-coordination.md) を参照。

## 使い方

```bash
tools/taskctl/taskctl build          # イメージを作り直す（mise.toml を変更したとき）
tools/taskctl/taskctl start <task>   # worktree を作成してコンテナを起動する
tools/taskctl/taskctl status         # タスクとコンテナの状態を一覧する
tools/taskctl/taskctl stop <task>    # コンテナを削除する（worktree は残る）
```

- `<task>` は英小文字・数字・`-` の40文字以内
- worktree は `~/.taskctl/worktrees/<task>` に、`origin/main` から `task/<task>` ブランチで作成する。このディレクトリはホストのシェル・IDEで開かない

## テスト

```bash
bash tools/taskctl/test/unit.sh              # 単体テスト（Docker 不要）
bash tools/taskctl/test/verify-container.sh  # 実コンテナでの隔離・通信制限の確認（Docker 必須）
```

`verify-container.sh` は、`CLAUDE_CODE_OAUTH_TOKEN` を設定して実行すると `claude -p` の動作も確認する。
