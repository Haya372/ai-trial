#!/bin/bash
# 通信制限なしでサブエージェントが動く状態を作らないため、
# ファイアウォールの設定に失敗したらコンテナを終了する（ADR-035 決定1）。
set -euo pipefail

if ! /usr/local/lib/taskctl/firewall.sh; then
  echo "taskctl: firewall setup failed; refusing to start" >&2
  exit 1
fi

# taskctl start はこのファイルの存在で起動完了を判定する
mkdir -p /run/taskctl
touch /run/taskctl/ready

exec sleep infinity
