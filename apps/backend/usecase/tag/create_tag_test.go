package tag_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Haya372/ai-trial/backend/domain"
	domaintag "github.com/Haya372/ai-trial/backend/domain/tag"
	tagmock "github.com/Haya372/ai-trial/backend/domain/tag/generated"
	taguc "github.com/Haya372/ai-trial/backend/usecase/tag"
)

const (
	fieldName    = "name"
	fieldColor   = "color"
	codeRequired = "REQUIRED"
	validName    = "work"
	validColor   = "red"
)

var errDBFailure = errors.New("db error")

type inTxKey struct{}

// markingTxManager marks the context passed to fn, so a repository mock can
// assert that it was called inside RunInTx.
type markingTxManager struct{}

func (markingTxManager) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(context.WithValue(ctx, inTxKey{}, true))
}

func inTx(ctx context.Context) bool {
	marked, _ := ctx.Value(inTxKey{}).(bool)
	return marked
}

// assertInTx fails when a repository call escapes the transaction that
// serializes the MaxTagsPerUser check (ADR-024).
func assertInTx(t *testing.T, ctx context.Context, method string) {
	t.Helper()
	if !inTx(ctx) {
		t.Errorf("%s must be called inside RunInTx", method)
	}
}

func TestCreateTagCommand_Execute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := tagmock.NewMockRepository(ctrl)
	userID := uuid.New()

	var created domaintag.Tag
	gomock.InOrder(
		repo.EXPECT().LockOwner(gomock.Any(), userID).DoAndReturn(
			func(ctx context.Context, _ uuid.UUID) error {
				assertInTx(t, ctx, "LockOwner")
				return nil
			},
		),
		repo.EXPECT().CountByUserID(gomock.Any(), userID).DoAndReturn(
			func(ctx context.Context, _ uuid.UUID) (int, error) {
				assertInTx(t, ctx, "CountByUserID")
				return domaintag.MaxTagsPerUser - 1, nil
			},
		),
		repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, tg domaintag.Tag) (domaintag.Tag, error) {
				assertInTx(t, ctx, "Create")
				created = tg
				return tg, nil
			},
		),
	)

	cmd := taguc.NewCreateTagCommand(repo, markingTxManager{})
	got, err := cmd.Execute(context.Background(), userID, taguc.CreateTagInput{Name: "  Work ", Color: "blue"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if created.ID() == uuid.Nil {
		t.Error("expected a generated tag ID, got uuid.Nil")
	}
	if created.UserID() != userID {
		t.Errorf("UserID: got %v, want %v", created.UserID(), userID)
	}
	if created.Name() != "Work" {
		t.Errorf("Name: got %q, want trimmed %q", created.Name(), "Work")
	}
	if created.Color() != domaintag.ColorBlue {
		t.Errorf("Color: got %q, want %q", created.Color(), domaintag.ColorBlue)
	}
	if got != created {
		t.Error("expected the tag returned by the repository")
	}
}

func TestCreateTagCommand_Execute_ValidationErrors(t *testing.T) {
	cases := []struct {
		name  string
		in    taguc.CreateTagInput
		wants []domain.ValidationDetail
	}{
		{
			name:  "empty name",
			in:    taguc.CreateTagInput{Name: "", Color: validColor},
			wants: []domain.ValidationDetail{{Field: fieldName, Code: codeRequired}},
		},
		{
			name:  "blank name",
			in:    taguc.CreateTagInput{Name: "   ", Color: validColor},
			wants: []domain.ValidationDetail{{Field: fieldName, Code: codeRequired}},
		},
		{
			name:  "name too long",
			in:    taguc.CreateTagInput{Name: strings.Repeat("a", 21), Color: validColor},
			wants: []domain.ValidationDetail{{Field: fieldName, Code: "TOO_LONG"}},
		},
		{
			name:  "unknown color",
			in:    taguc.CreateTagInput{Name: validName, Color: "black"},
			wants: []domain.ValidationDetail{{Field: fieldColor, Code: "INVALID_FORMAT"}},
		},
		{
			name: "name and color both invalid",
			in:   taguc.CreateTagInput{Name: "", Color: ""},
			wants: []domain.ValidationDetail{
				{Field: fieldName, Code: codeRequired},
				{Field: fieldColor, Code: "INVALID_FORMAT"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			// No repository call is expected: validation fails before the transaction.
			repo := tagmock.NewMockRepository(ctrl)

			cmd := taguc.NewCreateTagCommand(repo, markingTxManager{})
			_, err := cmd.Execute(context.Background(), uuid.New(), tc.in)

			ve, ok := errors.AsType[*domain.ValidationError](err)
			if !ok {
				t.Fatalf("expected *domain.ValidationError, got %T: %v", err, err)
			}
			if len(ve.Details) != len(tc.wants) {
				t.Fatalf("details: got %+v, want %+v", ve.Details, tc.wants)
			}
			for i, want := range tc.wants {
				if ve.Details[i].Field != want.Field || ve.Details[i].Code != want.Code {
					t.Errorf("details[%d]: got %s/%s, want %s/%s",
						i, ve.Details[i].Field, ve.Details[i].Code, want.Field, want.Code)
				}
			}
		})
	}
}

func TestCreateTagCommand_Execute_LimitReached_ReturnsErrTagLimitExceeded(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := tagmock.NewMockRepository(ctrl)
	userID := uuid.New()

	repo.EXPECT().LockOwner(gomock.Any(), userID).Return(nil)
	repo.EXPECT().CountByUserID(gomock.Any(), userID).Return(domaintag.MaxTagsPerUser, nil)
	// Create must not be called once the limit is reached.

	cmd := taguc.NewCreateTagCommand(repo, markingTxManager{})
	_, err := cmd.Execute(context.Background(), userID, taguc.CreateTagInput{Name: validName, Color: validColor})

	if !errors.Is(err, domaintag.ErrTagLimitExceeded) {
		t.Errorf("got %v, want ErrTagLimitExceeded", err)
	}
}

func TestCreateTagCommand_Execute_NameTaken_ReturnsErrTagNameTaken(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := tagmock.NewMockRepository(ctrl)
	userID := uuid.New()

	repo.EXPECT().LockOwner(gomock.Any(), userID).Return(nil)
	repo.EXPECT().CountByUserID(gomock.Any(), userID).Return(0, nil)
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, domaintag.ErrTagNameTaken)

	cmd := taguc.NewCreateTagCommand(repo, markingTxManager{})
	_, err := cmd.Execute(context.Background(), userID, taguc.CreateTagInput{Name: validName, Color: validColor})

	if !errors.Is(err, domaintag.ErrTagNameTaken) {
		t.Errorf("got %v, want ErrTagNameTaken", err)
	}
}

func TestCreateTagCommand_Execute_RepositoryErrors(t *testing.T) {
	cases := []struct {
		name   string
		expect func(repo *tagmock.MockRepository)
	}{
		{
			name: "LockOwner fails",
			expect: func(repo *tagmock.MockRepository) {
				repo.EXPECT().LockOwner(gomock.Any(), gomock.Any()).Return(errDBFailure)
			},
		},
		{
			name: "CountByUserID fails",
			expect: func(repo *tagmock.MockRepository) {
				repo.EXPECT().LockOwner(gomock.Any(), gomock.Any()).Return(nil)
				repo.EXPECT().CountByUserID(gomock.Any(), gomock.Any()).Return(0, errDBFailure)
			},
		},
		{
			name: "Create fails",
			expect: func(repo *tagmock.MockRepository) {
				repo.EXPECT().LockOwner(gomock.Any(), gomock.Any()).Return(nil)
				repo.EXPECT().CountByUserID(gomock.Any(), gomock.Any()).Return(0, nil)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errDBFailure)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := tagmock.NewMockRepository(ctrl)
			tc.expect(repo)

			cmd := taguc.NewCreateTagCommand(repo, markingTxManager{})
			_, err := cmd.Execute(context.Background(), uuid.New(), taguc.CreateTagInput{Name: validName, Color: validColor})

			if !errors.Is(err, errDBFailure) {
				t.Errorf("got %v, want wrapped errDBFailure", err)
			}
		})
	}
}
