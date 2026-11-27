#!/usr/bin/env bash
# push するコミット範囲だけを gitleaks でスキャンする。
# lefthook の pre-push から、git が渡す標準入力（push する ref の一覧）を受け取って実行する。
set -euo pipefail

while IFS=' ' read -r _local_ref local_sha _remote_ref remote_sha; do
    if [[ "$local_sha" == "0000000000000000000000000000000000000000" ]]; then
        continue
    fi

    if [[ "$remote_sha" == "0000000000000000000000000000000000000000" ]]; then
        log_opts="$local_sha"
    else
        log_opts="$remote_sha..$local_sha"
    fi

    gitleaks detect \
        --config .gitleaks.toml \
        --log-opts="$log_opts" \
        --verbose
done
