#!/usr/bin/env bash
# Go依存モジュールのうち、公開から MIN_RELEASE_AGE_DAYS 日経っていないバージョンを検出して失敗させる。
# Goには pnpm の minimumReleaseAge 相当の仕組みがないため、サプライチェーン攻撃対策として
# lefthook（pre-commit）と CI でこのチェックを実行する。
#
# 緊急のセキュリティ修正などで新しいバージョンを許可したい場合は、
# apps/backend/.go-release-age-exclude に「モジュールパス@バージョン」を1行ずつ記載する。
set -euo pipefail

MIN_RELEASE_AGE_DAYS="${MIN_RELEASE_AGE_DAYS:-7}"
BACKEND_DIR="$(cd "$(dirname "$0")/../apps/backend" && pwd)"
EXCLUDE_FILE="${BACKEND_DIR}/.go-release-age-exclude"

# BSD date（macOS）と GNU date（Linux）の両方に対応する
cutoff="$(date -u -v-"${MIN_RELEASE_AGE_DAYS}"d +%Y-%m-%dT%H:%M:%SZ 2>/dev/null \
  || date -u -d "${MIN_RELEASE_AGE_DAYS} days ago" +%Y-%m-%dT%H:%M:%SZ)"

excludes=""
if [[ -f "${EXCLUDE_FILE}" ]]; then
  excludes="$(grep -vE '^[[:space:]]*(#|$)' "${EXCLUDE_FILE}" || true)"
fi

modules="$(cd "${BACKEND_DIR}" && go list -m \
  -f '{{if not .Main}}{{with .Time}}{{.UTC.Format "2006-01-02T15:04:05Z"}} {{$.Path}}@{{$.Version}}{{end}}{{end}}' all)"

violations=""
while read -r published module; do
  [[ -z "${module:-}" ]] && continue
  if [[ "${published}" > "${cutoff}" ]] && ! grep -qxF "${module}" <<<"${excludes}"; then
    violations+="  ${module} (published ${published})"$'\n'
  fi
done <<<"${modules}"

if [[ -n "${violations}" ]]; then
  echo "error: 公開から${MIN_RELEASE_AGE_DAYS}日未満のGoモジュールが含まれています（基準: ${cutoff} 以前）" >&2
  printf '%s' "${violations}" >&2
  echo "${MIN_RELEASE_AGE_DAYS}日以上前に公開されたバージョンを指定してください。やむを得ない場合は apps/backend/.go-release-age-exclude に追記してください。" >&2
  exit 1
fi

echo "ok: すべてのGoモジュールが公開から${MIN_RELEASE_AGE_DAYS}日以上経過しています"
