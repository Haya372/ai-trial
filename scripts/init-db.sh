#!/bin/bash
set -euo pipefail

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<EOSQL
SELECT 'CREATE DATABASE "${DB_TEST_NAME:-ai_trial_test}"'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '${DB_TEST_NAME:-ai_trial_test}')\gexec
EOSQL
