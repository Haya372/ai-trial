#!/usr/bin/env bash
# taskctl が起動するコンテナを実際に動かし、Issue #303 の完了条件を確認する。
# Docker が必要。CLAUDE_CODE_OAUTH_TOKEN が設定されていれば claude -p の確認も行う。
# 実行: bash tools/taskctl/test/verify-container.sh
set -uo pipefail

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TASKCTL="$TEST_DIR/../taskctl"
TASK="verify-$$"
REAL_TOKEN="${CLAUDE_CODE_OAUTH_TOKEN:-}"
unset CLAUDE_CODE_OAUTH_TOKEN

# shellcheck source=/dev/null
source "$TASKCTL"
set +e

failures=0

check() {
  local name="$1"
  shift
  if "$@" >/dev/null 2>&1; then
    echo "ok   $name"
  else
    echo "FAIL $name"
    failures=$((failures + 1))
  fi
}

refute() {
  ! "$@"
}

in_container() {
  taskctl_exec "$TASK" "$@"
}

output_is() {
  local expected="$1"
  shift
  [ "$("$@")" = "$expected" ]
}

cap_is_zero() {
  in_container grep -qE "^$1:[[:space:]]*0+$" /proc/self/status
}

write_reaches_host() {
  in_container touch /workspace/.taskctl-write-test &&
    test -f "$WT/.taskctl-write-test" &&
    rm -f "$WT/.taskctl-write-test"
}

https_ok() {
  in_container curl -sS -o /dev/null --max-time 10 "$1"
}

exits_without_net_admin() {
  # 60秒以内に非0で終了すること（動き続けている場合は失敗とみなす）
  local id i code=""
  id="$(docker run -d --cap-drop ALL \
    --cap-add NET_RAW --cap-add NET_BIND_SERVICE \
    --cap-add SETUID --cap-add SETGID --cap-add SETPCAP \
    --security-opt no-new-privileges "$IMAGE")" || return 1
  for ((i = 0; i < 60; i++)); do
    if [ "$(docker inspect -f '{{.State.Running}}' "$id")" = "false" ]; then
      code="$(docker inspect -f '{{.State.ExitCode}}' "$id")"
      break
    fi
    sleep 1
  done
  docker rm -f "$id" >/dev/null
  [ -n "$code" ] && [ "$code" -ne 0 ]
}

token_visible_during_exec() {
  CLAUDE_CODE_OAUTH_TOKEN="dummy-token-$$" run_as_dev "$TASK" with-token \
    printenv CLAUDE_CODE_OAUTH_TOKEN </dev/null | grep -qx "dummy-token-$$"
}

token_in_resident_processes() {
  docker exec -u root "$(container_name "$TASK")" sh -c \
    'cat /proc/[0-9]*/environ 2>/dev/null | tr "\0" "\n"' | grep -q CLAUDE_CODE_OAUTH_TOKEN
}

token_in_container_config() {
  docker inspect "$(container_name "$TASK")" | grep -q CLAUDE_CODE_OAUTH_TOKEN
}

claude_runs_without_denials() {
  local result
  result="$(printf '%s\n' \
    "Create a file named claude-check.txt containing the word ok," \
    "then run 'cat claude-check.txt' with the Bash tool. Reply with DONE when finished." |
    CLAUDE_CODE_OAUTH_TOKEN="$REAL_TOKEN" taskctl_claude "$TASK" --output-format json)" || return 1
  echo "$result" >&2
  test -f "$WT/claude-check.txt" && grep -q '"permission_denials":\[\]' <<<"$result"
}

recreates_leftover_container() {
  docker stop -t 1 "$(container_name "$TASK")" &&
    bash "$TASKCTL" start "$TASK" &&
    in_container test -f /run/taskctl/ready
}

refuses_running_task() {
  ! bash "$TASKCTL" start "$TASK"
}

cleanup() {
  docker rm -f "$(container_name "$TASK")" >/dev/null 2>&1
  host_git worktree remove --force "$(worktree_path "$TASK")" >/dev/null 2>&1
  host_git branch -D "task/$TASK" >/dev/null 2>&1
}
trap cleanup EXIT

echo "# build & start"
bash "$TASKCTL" build >/dev/null || { echo "FAIL build"; exit 1; }
bash "$TASKCTL" start "$TASK" || { echo "FAIL start"; exit 1; }
WT="$(worktree_path "$TASK")"
check "start refuses a running task" refuses_running_task
check "start recreates a leftover stopped container" recreates_leftover_container

echo "# ホストのファイル・Docker socket が見えない"
check "host home is absent" in_container test ! -e "$HOME"
check "/Users is absent" in_container test ! -e /Users
check "repository checkout is absent" in_container test ! -e "$REPO_ROOT"
check "/var/run/docker.sock is absent" in_container test ! -e /var/run/docker.sock
check "/run/docker.sock is absent" in_container test ! -e /run/docker.sock
check "/workspace is the task worktree" in_container test -f /workspace/AGENTS.md
check "writes to /workspace reach the host" write_reaches_host

echo "# 一般ユーザー・capabilityなしで実行される"
check "runs as dev" output_is dev in_container id -un
check "no effective capabilities" cap_is_zero CapEff
check "empty bounding set" cap_is_zero CapBnd
check "cannot flush iptables" refute in_container iptables -F
check "cannot write /etc" refute in_container touch /etc/taskctl-test

echo "# 通信制限（暫定版。#304 でDNSフィルタリングに置き換える）"
check "api.anthropic.com:443 is reachable" https_ok https://api.anthropic.com
check "example.com is blocked" refute https_ok https://example.com
check "direct IP is blocked" refute https_ok https://1.1.1.1
check "allowed domain on port 80 is blocked" refute https_ok http://registry.npmjs.org

echo "# ファイアウォールを設定できなければ起動しない"
check "exits without NET_ADMIN" exits_without_net_admin

echo "# 認証トークンは実行中だけ存在する"
check "token is visible during exec" token_visible_during_exec
check "token is absent from resident processes" refute token_in_resident_processes
check "token is absent from container config" refute token_in_container_config

echo "# ビルド・テストが実行できる"
check "pnpm install" in_container pnpm install --frozen-lockfile
check "pnpm test (packages/ui)" in_container pnpm --filter @repo/ui test
# 生成コードは git 管理外のため、backend-ci.yml と同じ順で生成してからテストする
check "backend code generation" in_container bash -c '
  set -e
  pnpm --filter @ai-trial/schema bundle
  cd apps/backend
  make sqlc-generate oapi-codegen-generate
  go generate ./...'
check "backend make test" in_container make -C apps/backend test

echo "# claude -p"
if [ -n "$REAL_TOKEN" ]; then
  check "claude edits and runs Bash without permission denials" claude_runs_without_denials
else
  echo "skip claude -p (CLAUDE_CODE_OAUTH_TOKEN is not set)"
fi

echo
echo "failed: $failures"
[ "$failures" -eq 0 ]
