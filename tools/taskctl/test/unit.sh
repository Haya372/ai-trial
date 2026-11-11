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
  export FAKE_BRANCH_EXISTS=1
  # mise.toml の blob ハッシュ（タスクの worktree・origin/main・イメージのラベル）
  export FAKE_TASK_MISE=h1 FAKE_BASE_MISE=h1 FAKE_IMAGE_MISE=h1
  : >"$FAKE_LOG"

  # 呼び出しごとに "== <cmd>" の行と、引数を1行ずつ記録する
  cat >"$WORK/bin/docker" <<'EOF'
#!/usr/bin/env bash
{ echo "== docker"; printf '%s\n' "$@"; } >>"$FAKE_LOG"
case "$1" in
  image) echo "$FAKE_IMAGE_MISE" ;;
  ps) printf '%s' "${FAKE_STATE:-}" ;;
  inspect) echo "$FAKE_RUNNING" ;;
  exec)
    for a in "$@"; do
      if [ "$a" = "-i" ]; then cat >"$FAKE_STDIN"; fi
      if [ "$a" = "/usr/bin/test" ]; then exit "$FAKE_READY"; fi
    done
    ;;
esac
exit 0
EOF
  cat >"$WORK/bin/git" <<'EOF'
#!/usr/bin/env bash
{ echo "== git"; printf '%s\n' "$@"; } >>"$FAKE_LOG"
for a in "$@"; do
  if [ "$a" = "list" ]; then
    printf '%s' "$FAKE_WORKTREES"
    # 本物の git と同じく、SIGPIPE で書き込みに失敗したら非0で終了する
    if [ -n "${FAKE_WORKTREES_TAIL:-}" ]; then cat "$FAKE_WORKTREES_TAIL" || exit; fi
    exit 0
  fi
  if [ "$a" = "rev-parse" ]; then
    # --verify はブランチの存在確認。FAKE_BRANCH_EXISTS がその終了コード（0: ブランチあり）
    case " $* " in *" --verify "*) exit "$FAKE_BRANCH_EXISTS" ;; esac
    echo "$FAKE_BASE_MISE"
    exit 0
  fi
  if [ "$a" = "hash-object" ]; then echo "$FAKE_TASK_MISE"; exit 0; fi
  if [ "$a" = "show" ]; then echo "content of ${*: -1}"; exit 0; fi
  # 引数の最後から2番目がパス（... add [-b <branch>] <path> <commit-ish>）
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

# taskctl が使う worktree のパス。git に合わせてシンボリックリンクを解決している
# （macOS の /tmp は /private/tmp へのシンボリックリンク）
wt_path() {
  echo "$(cd "$HOME" && pwd -P)/.taskctl/worktrees/$1"
}

# 指定した引数を含む git の呼び出しを探して引数を返す
git_call() {
  local word="$1" i=1 args
  while :; do
    args="$(call_args git "$i")"
    [ -z "$args" ] && return 1
    if grep -qxF -- "$word" <<<"$args"; then
      printf '%s\n' "$args"
      return 0
    fi
    i=$((i + 1))
  done
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
  local wt
  wt="$(wt_path t1)"
  local args
  args="$(git_call add)" || fail "git worktree add not called"
  assert_pair "$args" "-c" "core.hooksPath=/dev/null"
  assert_pair "$args" "-c" "core.fsmonitor=false"
  assert_line "$args" "add"
  assert_pair "$args" "-b" "task/t1"
  assert_line "$args" "$wt"
}

test_start_reuses_registered_worktree() {
  local wt
  wt="$(wt_path t1)"
  mkdir -p "$wt"
  export FAKE_WORKTREES="worktree $wt"$'\n'
  bash "$TASKCTL" start t1 >/dev/null || fail "start failed"
  git_call add >/dev/null && fail "git worktree add was called"
  true
}

# 一致した行の後に出力が続くと、grep -q が先に終了して git が SIGPIPE で落ちうる
test_start_reuses_worktree_listed_before_long_output() {
  local wt i
  wt="$(wt_path t1)"
  mkdir -p "$wt"
  export FAKE_WORKTREES="worktree $wt"$'\n'
  export FAKE_WORKTREES_TAIL="$WORK/tail.txt"
  for ((i = 0; i < 20000; i++)); do echo "worktree /other/path-$i"; done >"$FAKE_WORKTREES_TAIL"
  bash "$TASKCTL" start t1 >/dev/null || fail "start failed"
  git_call add >/dev/null && fail "git worktree add was called"
  true
}

# git は worktree のパスをシンボリックリンクを解決して記録する（例: macOS の /tmp → /private/tmp）
test_start_reuses_worktree_when_home_is_a_symlink() {
  mkdir -p "$WORK/real-home"
  rm -rf "$HOME"
  ln -s "$WORK/real-home" "$HOME"
  local resolved
  resolved="$(cd "$WORK/real-home" && pwd -P)/.taskctl/worktrees/t1"
  mkdir -p "$resolved"
  export FAKE_WORKTREES="worktree $resolved"$'\n'
  bash "$TASKCTL" start t1 >/dev/null || fail "start failed"
  git_call add >/dev/null && fail "git worktree add was called"
  assert_pair "$(docker_call run)" "--mount" "type=bind,src=$resolved,dst=/workspace"
}

test_start_does_not_prune_when_worktree_exists() {
  local wt
  wt="$(wt_path t1)"
  mkdir -p "$wt"
  export FAKE_WORKTREES="worktree $wt"$'\n'
  bash "$TASKCTL" start t1 >/dev/null || fail "start failed"
  git_call prune >/dev/null && fail "git worktree prune was called"
  true
}

# ディレクトリを手で消すと、git には prunable な worktree として登録が残る
test_start_recreates_worktree_whose_directory_was_deleted() {
  local wt
  wt="$(wt_path t1)"
  export FAKE_WORKTREES="worktree $wt"$'\nprunable gitdir file points to non-existent location\n'
  bash "$TASKCTL" start t1 >/dev/null || fail "start failed"
  git_call prune >/dev/null || fail "git worktree prune not called"
  git_call add >/dev/null || fail "git worktree add not called"
  [ -d "$wt" ] || fail "worktree was not recreated"
}

# worktree を削除しても task/<task> ブランチは残るため、-b で作り直すと失敗する
test_start_reuses_leftover_branch() {
  export FAKE_BRANCH_EXISTS=0
  local wt args
  wt="$(wt_path t1)"
  bash "$TASKCTL" start t1 >/dev/null || fail "start failed"
  args="$(git_call add)" || fail "git worktree add not called"
  refute_line "$args" "-b"
  assert_sequence "$args" add "$wt" task/t1
}

test_start_refuses_unregistered_existing_directory() {
  mkdir -p "$(wt_path t1)"
  if bash "$TASKCTL" start t1 2>/dev/null; then
    fail "accepted a directory that is not a registered worktree"
  fi
  docker_call run >/dev/null && fail "docker run was called"
  true
}

test_start_runs_container_with_fixed_security_options() {
  bash "$TASKCTL" start t1 >/dev/null || fail "start failed"
  local args wt
  wt="$(wt_path t1)"
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

# root で実行するコマンドを PATH で解決すると、dev が書き込めるディレクトリの偽物が実行される
test_start_checks_readiness_with_absolute_path() {
  bash "$TASKCTL" start t1 >/dev/null || fail "start failed"
  local args
  args="$(docker_call exec)" || fail "docker exec not called"
  assert_sequence "$args" taskctl-t1 /usr/bin/test -f /run/taskctl/ready
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

# docker の呼び出しを、第1引数（サブコマンド）の並びとして返す
docker_subcommands() {
  awk '/^== / { take = ($0 == "== docker"); next } take { print; take = 0 }' "$FAKE_LOG"
}

test_start_removes_leftover_stopped_container_before_run() {
  export FAKE_STATE=exited
  bash "$TASKCTL" start t1 >/dev/null || fail "start failed"
  local args
  args="$(docker_call rm)" || fail "docker rm not called"
  assert_sequence "$args" rm -f taskctl-t1
  assert_sequence "$(docker_subcommands | grep -xE 'rm|run')" rm run
}

test_start_fails_when_task_is_already_running() {
  export FAKE_STATE=running
  local err
  if err="$(bash "$TASKCTL" start t1 2>&1 >/dev/null)"; then
    fail "start succeeded although the task is running"
  fi
  grep -q "already running" <<<"$err" || fail "unclear error: $err"
  docker_call run >/dev/null && fail "docker run was called"
  docker_call rm >/dev/null && fail "docker rm was called"
  true
}

# イメージのツールはビルド時の mise.toml で入るため、タスクの mise.toml と一致させる

test_start_does_not_build_when_image_matches_task_mise() {
  bash "$TASKCTL" start t1 >/dev/null || fail "start failed"
  docker_call build >/dev/null && fail "docker build was called"
  true
}

test_start_rebuilds_stale_image_when_task_matches_base() {
  export FAKE_TASK_MISE=h2 FAKE_BASE_MISE=h2 FAKE_IMAGE_MISE=h1
  bash "$TASKCTL" start t1 >/dev/null 2>&1 || fail "start failed"
  local args
  args="$(docker_call build)" || fail "docker build not called"
  assert_pair "$args" "--label" "taskctl.mise-toml=h2"
  assert_sequence "$(docker_subcommands | grep -xE 'build|run')" build run
}

test_start_fails_when_task_mise_differs_from_base() {
  export FAKE_TASK_MISE=h3 FAKE_BASE_MISE=h2 FAKE_IMAGE_MISE=h2
  local err
  if err="$(bash "$TASKCTL" start t1 2>&1 >/dev/null)"; then
    fail "start succeeded with a mismatched mise.toml"
  fi
  grep -q "mise.toml" <<<"$err" || fail "unclear error: $err"
  docker_call build >/dev/null && fail "docker build was called"
  docker_call run >/dev/null && fail "docker run was called"
  true
}

test_build_uses_mise_toml_from_base_ref() {
  bash "$TASKCTL" build >/dev/null || fail "build failed"
  local args ctx
  args="$(docker_call build)" || fail "docker build not called"
  assert_pair "$args" "--label" "taskctl.mise-toml=h1"
  git_call origin/main:mise.toml >/dev/null || fail "mise.toml was not read from origin/main"
  git_call origin/main:.env.example >/dev/null || fail ".env.example was not read from origin/main"
  # 作業ツリーのファイルではなく、origin/main から取り出したファイルをビルドに渡す
  ctx="$(awk 'prev == "--build-context" { print; exit } { prev = $0 }' <<<"$args")"
  [ "$ctx" != "project=$(cd "$TEST_DIR/../../.." && pwd)" ] || fail "built from the working tree"
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
  assert_sequence "$args" taskctl-t1 /usr/bin/setpriv --reuid=dev --regid=dev --init-groups \
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
