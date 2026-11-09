//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Haya372/ai-trial/backend/domain/tag"
	"github.com/Haya372/ai-trial/backend/infrastructure/db"
	"github.com/Haya372/ai-trial/backend/infrastructure/repository"
)

func newTestTag(t *testing.T, userID uuid.UUID, name, color string) tag.Tag {
	t.Helper()
	n, err := tag.NewTagName(name)
	if err != nil {
		t.Fatalf("build tag name: %v", err)
	}
	c, err := tag.NewTagColor(color)
	if err != nil {
		t.Fatalf("build tag color: %v", err)
	}
	return tag.New(uuid.New(), userID, n, c)
}

func createTestTag(t *testing.T, userID uuid.UUID, name string) tag.Tag {
	t.Helper()
	repo := repository.NewTagRepository(testPool, testLogger, testTracerProvider)
	created, err := repo.Create(context.Background(), newTestTag(t, userID, name, "red"))
	if err != nil {
		t.Fatalf("create test tag: %v", err)
	}
	return created
}

func mustTagName(t *testing.T, s string) tag.TagName {
	t.Helper()
	n, err := tag.NewTagName(s)
	if err != nil {
		t.Fatalf("build tag name: %v", err)
	}
	return n
}

// runInTestTx runs fn inside a transaction that is always rolled back, so
// a lock taken by fn is held until fn returns.
func runInTestTx(t *testing.T, fn func(ctx context.Context)) {
	t.Helper()
	err := db.NewPgxTxManager(testPool).RunInTx(context.Background(), func(ctx context.Context) error {
		fn(ctx)
		return errRollback
	})
	if !errors.Is(err, errRollback) {
		t.Fatalf("run in tx: %v", err)
	}
}

var errRollback = errors.New("rollback")

// isLockNotAvailable reports whether err is SQLSTATE 55P03, returned when
// NOWAIT or lock_timeout gives up on a row another transaction has locked.
func isLockNotAvailable(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "55P03"
}

func TestTagRepository_Create_PersistsTagAndReturnsIt(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	repo := repository.NewTagRepository(testPool, testLogger, testTracerProvider)

	want := newTestTag(t, u.ID(), "Work", "blue")
	got, err := repo.Create(context.Background(), want)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID() != want.ID() || got.UserID() != u.ID() {
		t.Errorf("id/userID mismatch: got %v/%v", got.ID(), got.UserID())
	}
	if string(got.Name()) != "Work" {
		t.Errorf("name mismatch: got %q", string(got.Name()))
	}
	if string(got.Color()) != "blue" {
		t.Errorf("color mismatch: got %q", string(got.Color()))
	}
}

func TestTagRepository_Create_SameNameIgnoringCase_ReturnsErrTagNameTaken(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	createTestTag(t, u.ID(), "Work")
	repo := repository.NewTagRepository(testPool, testLogger, testTracerProvider)

	_, err := repo.Create(context.Background(), newTestTag(t, u.ID(), "WORK", "red"))
	if !errors.Is(err, tag.ErrTagNameTaken) {
		t.Errorf("expected ErrTagNameTaken, got %v", err)
	}
}

func TestTagRepository_Create_SameNameForAnotherUser_Succeeds(t *testing.T) {
	setupTest(t)
	u1 := createTestUser(t)
	u2 := createTestUser(t)
	createTestTag(t, u1.ID(), "Work")
	repo := repository.NewTagRepository(testPool, testLogger, testTracerProvider)

	if _, err := repo.Create(context.Background(), newTestTag(t, u2.ID(), "Work", "red")); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTagRepository_FindByID_ReturnsTag(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	created := createTestTag(t, u.ID(), "Work")
	repo := repository.NewTagRepository(testPool, testLogger, testTracerProvider)

	got, err := repo.FindByID(context.Background(), created.ID())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID() != created.ID() || got.UserID() != u.ID() {
		t.Errorf("id/userID mismatch: got %v/%v", got.ID(), got.UserID())
	}
	if got.Name() != created.Name() || got.Color() != created.Color() {
		t.Errorf("name/color mismatch: got %q/%q", string(got.Name()), string(got.Color()))
	}
}

func TestTagRepository_FindByID_NotFound_ReturnsErrTagNotFound(t *testing.T) {
	setupTest(t)
	repo := repository.NewTagRepository(testPool, testLogger, testTracerProvider)

	_, err := repo.FindByID(context.Background(), uuid.New())
	if !errors.Is(err, tag.ErrTagNotFound) {
		t.Errorf("expected ErrTagNotFound, got %v", err)
	}
}

func TestTagRepository_Update_PersistsChanges(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	created := createTestTag(t, u.ID(), "Work")
	repo := repository.NewTagRepository(testPool, testLogger, testTracerProvider)

	color, _ := tag.NewTagColor("green")
	if err := repo.Update(context.Background(), tag.New(created.ID(), u.ID(), mustTagName(t, "Office"), color)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.FindByID(context.Background(), created.ID())
	if err != nil {
		t.Fatalf("find updated tag: %v", err)
	}
	if string(got.Name()) != "Office" || string(got.Color()) != "green" {
		t.Errorf("name/color mismatch: got %q/%q", string(got.Name()), string(got.Color()))
	}
}

func TestTagRepository_Update_OwnNameWithDifferentCase_Succeeds(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	created := createTestTag(t, u.ID(), "Work")
	repo := repository.NewTagRepository(testPool, testLogger, testTracerProvider)

	if err := repo.Update(context.Background(), tag.New(created.ID(), u.ID(), mustTagName(t, "WORK"), created.Color())); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTagRepository_Update_NameOfAnotherTag_ReturnsErrTagNameTaken(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	createTestTag(t, u.ID(), "Work")
	other := createTestTag(t, u.ID(), "Private")
	repo := repository.NewTagRepository(testPool, testLogger, testTracerProvider)

	err := repo.Update(context.Background(), tag.New(other.ID(), u.ID(), mustTagName(t, "work"), other.Color()))
	if !errors.Is(err, tag.ErrTagNameTaken) {
		t.Errorf("expected ErrTagNameTaken, got %v", err)
	}
}

func TestTagsTable_UpdatedAt_IsRefreshedByTrigger(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	created := createTestTag(t, u.ID(), "Work")
	ctx := context.Background()

	// Backdate with the trigger disabled, since the trigger would overwrite
	// the explicit value too.
	past := time.Now().Add(-time.Hour)
	if _, err := testPool.Exec(ctx, "ALTER TABLE tags DISABLE TRIGGER update_tags_updated_at"); err != nil {
		t.Fatalf("disable trigger: %v", err)
	}
	_, err := testPool.Exec(ctx, "UPDATE tags SET updated_at = $1 WHERE id = $2", past, created.ID())
	if _, enableErr := testPool.Exec(ctx, "ALTER TABLE tags ENABLE TRIGGER update_tags_updated_at"); enableErr != nil {
		t.Fatalf("enable trigger: %v", enableErr)
	}
	if err != nil {
		t.Fatalf("backdate updated_at without trigger: %v", err)
	}

	if _, err := testPool.Exec(ctx, "UPDATE tags SET name = 'Office' WHERE id = $1", created.ID()); err != nil {
		t.Fatalf("update tag: %v", err)
	}

	var updatedAt time.Time
	if err := testPool.QueryRow(ctx, "SELECT updated_at FROM tags WHERE id = $1", created.ID()).Scan(&updatedAt); err != nil {
		t.Fatalf("select updated_at: %v", err)
	}
	if !updatedAt.After(past.Add(time.Minute)) {
		t.Errorf("expected updated_at to be refreshed, got %v", updatedAt)
	}
}

func TestTagRepository_Delete_RemovesTag(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	created := createTestTag(t, u.ID(), "Work")
	repo := repository.NewTagRepository(testPool, testLogger, testTracerProvider)

	if err := repo.Delete(context.Background(), created.ID()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := repo.FindByID(context.Background(), created.ID()); !errors.Is(err, tag.ErrTagNotFound) {
		t.Errorf("expected ErrTagNotFound after delete, got %v", err)
	}
}

func TestTagRepository_Delete_RemovesAssignmentsButKeepsEvents(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	created := createTestTag(t, u.ID(), "Work")
	now := time.Now().UTC()
	eventID := insertEvent(t, u.ID(), "Meeting", now, now.Add(time.Hour))
	insertEventTag(t, eventID, created.ID(), 1)
	repo := repository.NewTagRepository(testPool, testLogger, testTracerProvider)

	if err := repo.Delete(context.Background(), created.ID()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n := countRows(t, "SELECT COUNT(*) FROM event_tags WHERE event_id = $1", eventID); n != 0 {
		t.Errorf("expected assignments to be deleted, got %d", n)
	}
	if n := countRows(t, "SELECT COUNT(*) FROM events WHERE id = $1", eventID); n != 1 {
		t.Errorf("expected event to remain, got %d", n)
	}
}

func TestEventTagsTable_EventDeletion_RemovesAssignmentsButKeepsTags(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	created := createTestTag(t, u.ID(), "Work")
	now := time.Now().UTC()
	eventID := insertEvent(t, u.ID(), "Meeting", now, now.Add(time.Hour))
	insertEventTag(t, eventID, created.ID(), 1)

	if _, err := testPool.Exec(context.Background(), "DELETE FROM events WHERE id = $1", eventID); err != nil {
		t.Fatalf("delete event: %v", err)
	}

	if n := countRows(t, "SELECT COUNT(*) FROM event_tags WHERE tag_id = $1", created.ID()); n != 0 {
		t.Errorf("expected assignments to be deleted, got %d", n)
	}
	if n := countRows(t, "SELECT COUNT(*) FROM tags WHERE id = $1", created.ID()); n != 1 {
		t.Errorf("expected tag to remain, got %d", n)
	}
}

func TestEventTagsTable_RejectsDuplicateTagAndPosition(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	work := createTestTag(t, u.ID(), "Work")
	private := createTestTag(t, u.ID(), "Private")
	now := time.Now().UTC()
	eventID := insertEvent(t, u.ID(), "Meeting", now, now.Add(time.Hour))
	insertEventTag(t, eventID, work.ID(), 1)

	tests := []struct {
		name     string
		tagID    uuid.UUID
		position int
	}{
		{"same tag twice", work.ID(), 2},
		{"same position twice", private.ID(), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testPool.Exec(context.Background(),
				"INSERT INTO event_tags (event_id, tag_id, position) VALUES ($1, $2, $3)",
				eventID, tt.tagID, tt.position,
			)
			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
				t.Errorf("expected unique violation, got %v", err)
			}
		})
	}
}

func TestEventTagsTable_HasTagIDIndex(t *testing.T) {
	setupTest(t)
	n := countRows(t, "SELECT COUNT(*) FROM pg_indexes WHERE tablename = 'event_tags' AND indexname = 'idx_event_tags_tag_id'")
	if n != 1 {
		t.Errorf("expected idx_event_tags_tag_id to exist, got %d", n)
	}
}

func TestTagRepository_CountByUserID_CountsOnlyOwnTags(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	other := createTestUser(t)
	createTestTag(t, u.ID(), "Work")
	createTestTag(t, u.ID(), "Private")
	createTestTag(t, other.ID(), "Work")
	repo := repository.NewTagRepository(testPool, testLogger, testTracerProvider)

	got, err := repo.CountByUserID(context.Background(), u.ID())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 2 {
		t.Errorf("expected 2, got %d", got)
	}
}

func TestTagRepository_LockOwner_LocksUserRowUntilTxEnds(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	repo := repository.NewTagRepository(testPool, testLogger, testTracerProvider)

	runInTestTx(t, func(ctx context.Context) {
		if err := repo.LockOwner(ctx, u.ID()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err := testPool.Exec(context.Background(), "SELECT id FROM users WHERE id = $1 FOR UPDATE NOWAIT", u.ID())
		if !isLockNotAvailable(err) {
			t.Errorf("expected the user row to be locked, got %v", err)
		}
	})
}

func TestTagRepository_LockOwner_OutsideTx_ReturnsError(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	repo := repository.NewTagRepository(testPool, testLogger, testTracerProvider)

	if err := repo.LockOwner(context.Background(), u.ID()); err == nil {
		t.Error("expected an error outside a transaction")
	}
}

func TestTagRepository_FindByIDsForShare_ReturnsExistingTagsOnly(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	work := createTestTag(t, u.ID(), "Work")
	private := createTestTag(t, u.ID(), "Private")
	repo := repository.NewTagRepository(testPool, testLogger, testTracerProvider)

	runInTestTx(t, func(ctx context.Context) {
		got, err := repo.FindByIDsForShare(ctx, []uuid.UUID{work.ID(), uuid.New(), private.ID()})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		ids := map[uuid.UUID]tag.Tag{}
		for _, tg := range got {
			ids[tg.ID()] = tg
		}
		if len(got) != 2 || ids[work.ID()] == nil || ids[private.ID()] == nil {
			t.Errorf("expected work and private tags, got %d tags", len(got))
		}
		if ids[work.ID()] != nil && ids[work.ID()].UserID() != u.ID() {
			t.Errorf("userID mismatch: got %v", ids[work.ID()].UserID())
		}
	})
}

func TestTagRepository_FindByIDsForShare_EmptyIDs_ReturnsEmpty(t *testing.T) {
	setupTest(t)
	repo := repository.NewTagRepository(testPool, testLogger, testTracerProvider)

	runInTestTx(t, func(ctx context.Context) {
		got, err := repo.FindByIDsForShare(ctx, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("expected no tags, got %d", len(got))
		}
	})
}

func TestTagRepository_FindByIDsForShare_BlocksDeletionUntilTxEnds(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	work := createTestTag(t, u.ID(), "Work")
	repo := repository.NewTagRepository(testPool, testLogger, testTracerProvider)

	runInTestTx(t, func(ctx context.Context) {
		if _, err := repo.FindByIDsForShare(ctx, []uuid.UUID{work.ID()}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		err := pgx.BeginFunc(context.Background(), testPool, func(other pgx.Tx) error {
			if _, err := other.Exec(context.Background(), "SET LOCAL lock_timeout = '200ms'"); err != nil {
				return err
			}
			_, err := other.Exec(context.Background(), "DELETE FROM tags WHERE id = $1", work.ID())
			return err
		})
		if !isLockNotAvailable(err) {
			t.Errorf("expected the tag row to be locked against deletion, got %v", err)
		}
	})
}

func TestTagRepository_FindByIDsForShare_OutsideTx_ReturnsError(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	work := createTestTag(t, u.ID(), "Work")
	repo := repository.NewTagRepository(testPool, testLogger, testTracerProvider)

	if _, err := repo.FindByIDsForShare(context.Background(), []uuid.UUID{work.ID()}); err == nil {
		t.Error("expected an error outside a transaction")
	}
}

func insertEventTag(t *testing.T, eventID, tagID uuid.UUID, position int) {
	t.Helper()
	_, err := testPool.Exec(context.Background(),
		"INSERT INTO event_tags (event_id, tag_id, position) VALUES ($1, $2, $3)", eventID, tagID, position,
	)
	if err != nil {
		t.Fatalf("insert event tag: %v", err)
	}
}

func countRows(t *testing.T, sql string, args ...any) int {
	t.Helper()
	var n int
	if err := testPool.QueryRow(context.Background(), sql, args...).Scan(&n); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	return n
}
