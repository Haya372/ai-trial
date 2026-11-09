#!/usr/bin/env bash
# taskctl の単体テスト。docker / git を偽物に差し替え、渡される引数を検証する。
# 実行: bash tools/taskctl/test/unit.sh
set -uo pipefail

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TASKCTL="$TEST_DIR/../taskctl"

failures=0
passes=0

setup() {
  # macOS の mktemp はテンプレートなしだと TMPDIR を見ないため明示する
  WORK="$(mktemp -d "${TMPDIR:-/tmp}/taskctl-test.XXXXXX")" || exit 1
  export HOME="$WORK/home"
  mkdir -p "$HOME" "$WORK/bin"
  export FAKE_LOG="$WORK/calls.log"
  export FAKE_STDIN="$WORK/stdin.log"
  export FAKE_RUNNING=true
  export FAKE_READY=0
  export FAKE_WORKTREES=""
  : >"$FAKE_LOG"

  # 呼び出しごとに "== <cmd>" の行と、引数を1行ずつ記録する
  cat >"$WORK/bin/docker" <<'EOF'
#!/usr/bin/env bash
{ echo "== docker"; printf '%s\n' "$@"; } >>"$FAKE_LOG"
case "$1" in
  image) exit 0 ;;
  inspect) echo "$FAKE_RUNNING" ;;
  exec)
    for a in "$@"; do
      if [ "$a" = "-i" ]; then cat >"$FAKE_STDIN"; fi
      if [ "$a" = "test" ]; then exit "$FAKE_READY"; fi
    done
    ;;
esac
exit 0
EOF
  cat >"$WORK/bin/git" <<'EOF'
#!/usr/bin/env bash
{ echo "== git"; printf '%s\n' "$@"; } >>"$FAKE_LOG"
for a in "$@"; do
  if [ "$a" = "list" ]; then printf '%s' "$FAKE_WORKTREES"; exit 0; fi
  # 引数の最後から2番目がパス（... add -b <branch> <path> <commit>）
  if [ "$a" = "add" ]; then mkdir -p "${@: -2:1}"; fi
done
exit 0
EOF
  chmod +x "$WORK/bin/docker" "$WORK/bin/git"
  export PATH="$WORK/bin:$PATH"
  unset CLAUDE_CODE_OAUTH_TOKEN
}

teardown() {
  rm -rf "$WORK"
}

# 記録の中から、指定コマンドの n 番目（1始まり）の呼び出しの引数を取り出す
call_args() {
  local cmd="$1" n="$2"
  awk -v cmd="== $cmd" -v n="$n" '
    /^== / { if ($0 == cmd) { i++ } inblock = ($0 == cmd && i == n); next }
    inblock { print }
  ' "$FAKE_LOG"
}

# 指定サブコマンド（docker の第1引数）の呼び出しを探して引数を返す
docker_call() {
  local sub="$1" i=1 args
  while :; do
    args="$(call_args docker "$i")"
    [ -z "$args" ] && return 1
    if [ "$(head -n1 <<<"$args")" = "$sub" ]; then
      printf '%s\n' "$args"
      return 0
    fi
    i=$((i + 1))
  done
}

fail() {
  echo "    FAIL: $*" >&2
  return 1
}

assert_line() {
  local haystack="$1" needle="$2"
  grep -qxF -- "$needle" <<<"$haystack" || fail "missing line: $needle"
}

refute_line() {
  local haystack="$1" needle="$2"
  ! grep -qxF -- "$needle" <<<"$haystack" || fail "unexpected line: $needle"
}

refute_contains() {
  local haystack="$1" needle="$2"
  ! grep -qF -- "$needle" <<<"$haystack" || fail "unexpected text: $needle"
}

# 連続する行（オプションと値）を検証する
assert_pair() {
  local haystack="$1" opt="$2" val="$3"
  awk -v o="$opt" -v v="$val" 'prev == o && $0 == v { found = 1 } { prev = $0 } END { exit !found }' \
    <<<"$haystack" || fail "missing pair: $opt $val"
}

# 引数の並び全体が連続して現れることを検証する（grep -F に改行入りのパターンを渡すと
# 行ごとの OR になり並びを検証できないため、awk で照合する）
assert_sequence() {
  local haystack="$1"
  shift
  awk -v seq="$(printf '%s\037' "$@")" '
    BEGIN { n = split(seq, e, "\037") - 1 }
    { l[NR] = $0 }
    END {
      for (i = 1; i + n - 1 <= NR; i++) {
        ok = 1
        for (j = 1; j <= n; j++) if (l[i + j - 1] != e[j]) { ok = 0; break }
        if (ok) exit 0
      }
      exit 1
    }' <<<"$haystack" || fail "missing sequence: $*"
}

count_pairs() {
  local haystack="$1" opt="$2"
  grep -cxF -- "$opt" <<<"$haystack"
}

run_test() {
  local name="$1"
  local rc
  setup
  # if の条件や || の左辺では set -e が無効になるため、単独の文として実行する
  ( set -e; "$name" )
  rc=$?
  if [ "$rc" -eq 0 ]; then
    passes=$((passes + 1))
    echo "ok   $name"
  else
    failures=$((failures + 1))
    echo "FAIL $name"
  fi
  teardown
}

# ---- テスト用ヘルパー ----

test_assert_sequence_detects_order_and_gaps() {
  local lines
  lines="$(printf '%s\n' a b c d)"
  assert_sequence "$lines" b c
  ! (assert_sequence "$lines" c b 2>/dev/null) || fail "accepted reversed order"
  ! (assert_sequence "$lines" a c 2>/dev/null) || fail "accepted non-adjacent lines"
  ! (assert_sequence "$lines" d e 2>/dev/null) || fail "accepted a sequence past the end"
}

# ---- start ----

test_start_rejects_invalid_task_names() {
  local name
  for name in "" "Upper" "../x" "a b" "-x" "a/b" "a;b" "$(printf 'a%.0s' {1..41})"; do
    if bash "$TASKCTL" start "$name" 2>/dev/null; then
      fail "accepted invalid name: '$name'"
    fi
  done
  [ ! -s "$FAKE_LOG" ] || fail "docker/git was called for invalid names"
}

test_start_accepts_boundary_task_names() {
  bash "$TASKCTL" start "a" >/dev/null || fail "rejected 'a'"
  bash "$TASKCTL" start "$(printf 'a%.0s' {1..40})" >/dev/null || fail "rejected 40 chars"
}

test_start_rejects_extra_arguments() {
  if bash "$TASKCTL" start t1 -v /:/host 2>/dev/null; then
    fail "accepted extra arguments"
  fi
  [ ! -s "$FAKE_LOG" ] || fail "docker/git was called"
}

test_start_creates_worktree_with_hooks_disabled() {
  bash "$TASKCTL" start t1 >/dev/null || fail "start failed"
  local wt="$HOME/.taskctl/worktrees/t1"
  local args
  args="$(call_args git 2)"
  assert_pair "$args" "-c" "core.hooksPath=/dev/null"
  assert_pair "$args" "-c" "core.fsmonitor=false"
  assert_line "$args" "add"
  assert_pair "$args" "-b" "task/t1"
  assert_line "$args" "$wt"
}

test_start_reuses_registered_worktree() {
  local wt="$HOME/.taskctl/worktrees/t1"
  mkdir -p "$wt"
  export FAKE_WORKTREES="worktree $wt"$'\n'
  bash "$TASKCTL" start t1 >/dev/null || fail "start failed"
  refute_line "$(call_args git 2)" "add"
}

test_start_refuses_unregistered_existing_directory() {
  mkdir -p "$HOME/.taskctl/worktrees/t1"
  if bash "$TASKCTL" start t1 2>/dev/null; then
    fail "accepted a directory that is not a registered worktree"
  fi
  docker_call run >/dev/null && fail "docker run was called"
  true
}

test_start_runs_container_with_fixed_security_options() {
  bash "$TASKCTL" start t1 >/dev/null || fail "start failed"
  local args wt="$HOME/.taskctl/worktrees/t1"
  args="$(docker_call run)" || fail "docker run not called"
  assert_pair "$args" "--name" "taskctl-t1"
  assert_pair "$args" "--cap-drop" "ALL"
  assert_pair "$args" "--security-opt" "no-new-privileges"
  local cap
  for cap in NET_ADMIN NET_RAW NET_BIND_SERVICE SETUID SETGID SETPCAP; do
    assert_pair "$args" "--cap-add" "$cap"
  done
  [ "$(count_pairs "$args" "--cap-add")" -eq 6 ] || fail "unexpected number of --cap-add"
  [ "$(count_pairs "$args" "--mount")" -eq 1 ] || fail "expected exactly one --mount"
  assert_pair "$args" "--mount" "type=bind,src=$wt,dst=/workspace"
  refute_line "$args" "-v"
  refute_line "$args" "--volume"
  refute_line "$args" "--privileged"
  refute_contains "$args" "docker.sock"
  refute_contains "$args" "CLAUDE_CODE_OAUTH_TOKEN"
}

test_start_does_not_pass_token_to_docker_run() {
  CLAUDE_CODE_OAUTH_TOKEN=secret-token bash "$TASKCTL" start t1 >/dev/null || fail "start failed"
  refute_contains "$(cat "$FAKE_LOG")" "secret-token"
  refute_contains "$(docker_call run)" "-e"
}

test_start_fails_and_removes_container_when_entrypoint_exits() {
  export FAKE_RUNNING=false
  if bash "$TASKCTL" start t1 >/dev/null 2>&1; then
    fail "start succeeded although the container exited"
  fi
  local args
  args="$(docker_call rm)" || fail "docker rm not called"
  assert_line "$args" "taskctl-t1"
}

# ---- stop / status ----

test_stop_removes_container() {
  bash "$TASKCTL" stop t1 >/dev/null || fail "stop failed"
  local args
  args="$(docker_call rm)" || fail "docker rm not called"
  assert_line "$args" "-f"
  assert_line "$args" "taskctl-t1"
}

test_stop_rejects_invalid_task_name() {
  if bash "$TASKCTL" stop "../x" 2>/dev/null; then
    fail "accepted invalid name"
  fi
  [ ! -s "$FAKE_LOG" ] || fail "docker was called"
}

test_status_lists_task_containers_only() {
  bash "$TASKCTL" status >/dev/null || fail "status failed"
  local args
  args="$(docker_call ps)" || fail "docker ps not called"
  assert_pair "$args" "--filter" "label=taskctl.task"
}

test_unknown_subcommand_fails_with_usage() {
  local status=0
  bash "$TASKCTL" run t1 2>/dev/null || status=$?
  [ "$status" -eq 2 ] || fail "expected exit 2, got $status"
  [ ! -s "$FAKE_LOG" ] || fail "docker was called"
}

# ---- exec / claude（ライブラリとして読み込んで検証する） ----

test_exec_drops_to_dev_without_capabilities() {
  # shellcheck source=/dev/null
  source "$TASKCTL"
  taskctl_exec t1 echo hello || fail "taskctl_exec failed"
  local args
  args="$(docker_call exec)" || fail "docker exec not called"
  assert_sequence "$args" exec -u root -w /workspace
  assert_sequence "$args" taskctl-t1 setpriv --reuid=dev --regid=dev --init-groups \
    --inh-caps=-all --bounding-set=-all -- echo hello
  [ "$(tail -n1 <<<"$args")" = hello ] || fail "command is not at the end"
}

test_exec_does_not_accept_docker_options() {
  # shellcheck source=/dev/null
  source "$TASKCTL"
  taskctl_exec t1 --privileged -e LD_PRELOAD=x id || fail "taskctl_exec failed"
  local args
  args="$(docker_call exec)" || fail "docker exec not called"
  assert_sequence "$args" exec -u root
  assert_sequence "$args" -- --privileged -e LD_PRELOAD=x id
  [ "$(grep -cxF -- --privileged <<<"$args")" -eq 1 ] || fail "--privileged passed to docker"
}

test_exec_rejects_invalid_task_name() {
  # shellcheck source=/dev/null
  source "$TASKCTL"
  if (taskctl_exec "../x" id 2>/dev/null); then
    fail "accepted invalid name"
  fi
  [ ! -s "$FAKE_LOG" ] || fail "docker was called"
}

# シェルに展開されないことを確かめるため、あえて展開されうる文字列を single quote で渡す
# shellcheck disable=SC2016
test_exec_passes_arguments_verbatim() {
  # shellcheck source=/dev/null
  source "$TASKCTL"
  taskctl_exec t1 printf '%s' '$(id)' 'a b' '*' || fail "taskctl_exec failed"
  local args
  args="$(docker_call exec)"
  assert_line "$args" '$(id)'
  assert_line "$args" 'a b'
  assert_line "$args" '*'
}

test_claude_requires_token() {
  # shellcheck source=/dev/null
  source "$TASKCTL"
  # die は exit するため、サブシェルで呼ぶ
  if (taskctl_claude t1 </dev/null 2>/dev/null); then
    fail "succeeded without CLAUDE_CODE_OAUTH_TOKEN"
  fi
  docker_call exec >/dev/null && fail "docker exec was called"
  true
}

test_claude_passes_token_by_name_and_prompt_by_stdin() {
  # shellcheck source=/dev/null
  source "$TASKCTL"
  export CLAUDE_CODE_OAUTH_TOKEN=secret-token
  echo "do the task" | taskctl_claude t1 --output-format json || fail "taskctl_claude failed"
  local args
  args="$(docker_call exec)" || fail "docker exec not called"
  assert_sequence "$args" exec -i -e CLAUDE_CODE_OAUTH_TOKEN -u root
  refute_contains "$(cat "$FAKE_LOG")" "secret-token"
  assert_sequence "$args" -- claude -p --permission-mode acceptEdits --allowedTools Bash \
    --output-format json
  [ "$(tail -n1 <<<"$args")" = json ] || fail "extra arguments are not at the end"
  [ "$(cat "$FAKE_STDIN")" = "do the task" ] || fail "prompt was not passed via stdin"
}

main() {
  local t
  for t in $(declare -F | awk '{print $3}' | grep '^test_'); do
    run_test "$t"
  done
  echo
  echo "passed: $passes, failed: $failures"
  [ "$failures" -eq 0 ]
}

main
