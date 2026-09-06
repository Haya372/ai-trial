ifneq (,$(wildcard .env))
  include .env
  export
endif

.PHONY: db/up db/down db/logs db/reset db/psql

db/up:
	docker compose up -d --wait postgres

db/down:
	docker compose rm -sf postgres

db/logs:
	docker compose logs -f postgres

db/reset:
	@echo "WARNING: This will destroy all data in the postgres volume. Ctrl+C to cancel."
	@sleep 3
	docker compose down -v

db/psql:
	@docker compose ps postgres --status running --quiet 2>/dev/null | grep -q . || \
		{ echo "ERROR: postgres is not running. Run 'make db/up' first." >&2; exit 1; }
	docker compose exec postgres psql -U $${DB_USER:-postgres} $${DB_NAME:-ai_trial_dev}
