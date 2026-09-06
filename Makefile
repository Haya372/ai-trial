.PHONY: db/up db/down db/logs db/reset db/psql

db/up:
	docker compose up -d --wait postgres

db/down:
	docker compose rm -sf postgres

db/logs:
	docker compose logs -f postgres

db/reset:
	docker compose down -v

db/psql:
	docker compose exec postgres psql -U $${DB_USER:-postgres} $${DB_NAME:-ai_trial_dev}
