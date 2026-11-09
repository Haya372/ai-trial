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

func newStoredTag(t *testing.T, id, userID uuid.UUID) domaintag.Tag {
	t.Helper()
	name, err := domaintag.NewTagName("Old")
	if err != nil {
		t.Fatalf("build tag name: %v", err)
	}
	color, err := domaintag.NewTagColor("green")
	if err != nil {
		t.Fatalf("build tag color: %v", err)
	}
	return domaintag.New(id, userID, name, color)
}

func TestUpdateTagCommand_Execute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := tagmock.NewMockRepository(ctrl)
	userID := uuid.New()
	tagID := uuid.New()

	var updated domaintag.Tag
	gomock.InOrder(
		repo.EXPECT().FindByID(gomock.Any(), tagID).DoAndReturn(
			func(ctx context.Context, _ uuid.UUID) (domaintag.Tag, error) {
				assertInTx(t, ctx, "FindByID")
				return newStoredTag(t, tagID, userID), nil
			},
		),
		repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, tg domaintag.Tag) error {
				assertInTx(t, ctx, "Update")
				updated = tg
				return nil
			},
		),
		repo.EXPECT().CountEventsByTagID(gomock.Any(), tagID).DoAndReturn(
			func(ctx context.Context, _ uuid.UUID) (int, error) {
				assertInTx(t, ctx, "CountEventsByTagID")
				return 3, nil
			},
		),
	)

	cmd := taguc.NewUpdateTagCommand(repo, markingTxManager{})
	got, eventCount, err := cmd.Execute(context.Background(), userID, taguc.UpdateTagInput{
		ID: tagID, Name: "  Work ", Color: otherColor,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.ID() != tagID {
		t.Errorf("ID: got %v, want %v", updated.ID(), tagID)
	}
	if updated.UserID() != userID {
		t.Errorf("UserID: got %v, want %v", updated.UserID(), userID)
	}
	if updated.Name() != "Work" {
		t.Errorf("Name: got %q, want trimmed %q", updated.Name(), "Work")
	}
	if updated.Color() != domaintag.ColorBlue {
		t.Errorf("Color: got %q, want %q", updated.Color(), domaintag.ColorBlue)
	}
	if got != updated {
		t.Error("expected the tag passed to the repository")
	}
	if eventCount != 3 {
		t.Errorf("eventCount: got %d, want 3", eventCount)
	}
}

func TestUpdateTagCommand_Execute_ValidationErrors(t *testing.T) {
	cases := []struct {
		name  string
		in    taguc.UpdateTagInput
		wants []domain.ValidationDetail
	}{
		{
			name:  "empty name",
			in:    taguc.UpdateTagInput{ID: uuid.New(), Name: "", Color: validColor},
			wants: []domain.ValidationDetail{{Field: fieldName, Code: codeRequired}},
		},
		{
			name:  "blank name",
			in:    taguc.UpdateTagInput{ID: uuid.New(), Name: "   ", Color: validColor},
			wants: []domain.ValidationDetail{{Field: fieldName, Code: codeRequired}},
		},
		{
			name:  "name too long",
			in:    taguc.UpdateTagInput{ID: uuid.New(), Name: strings.Repeat("a", 21), Color: validColor},
			wants: []domain.ValidationDetail{{Field: fieldName, Code: "TOO_LONG"}},
		},
		{
			name:  "unknown color",
			in:    taguc.UpdateTagInput{ID: uuid.New(), Name: validName, Color: "black"},
			wants: []domain.ValidationDetail{{Field: fieldColor, Code: codeInvalidFormat}},
		},
		{
			name: "name and color both invalid",
			in:   taguc.UpdateTagInput{ID: uuid.New(), Name: "", Color: ""},
			wants: []domain.ValidationDetail{
				{Field: fieldName, Code: codeRequired},
				{Field: fieldColor, Code: codeInvalidFormat},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			// No repository call is expected: validation fails before the transaction.
			repo := tagmock.NewMockRepository(ctrl)

			cmd := taguc.NewUpdateTagCommand(repo, markingTxManager{})
			_, _, err := cmd.Execute(context.Background(), uuid.New(), tc.in)

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

func TestUpdateTagCommand_Execute_TagNotFound_ReturnsErrTagNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := tagmock.NewMockRepository(ctrl)
	tagID := uuid.New()

	repo.EXPECT().FindByID(gomock.Any(), tagID).Return(nil, domaintag.ErrTagNotFound)

	cmd := taguc.NewUpdateTagCommand(repo, markingTxManager{})
	_, _, err := cmd.Execute(context.Background(), uuid.New(), taguc.UpdateTagInput{
		ID: tagID, Name: validName, Color: validColor,
	})

	if !errors.Is(err, domaintag.ErrTagNotFound) {
		t.Errorf("got %v, want ErrTagNotFound", err)
	}
}

func TestUpdateTagCommand_Execute_AnotherUsersTag_ReturnsErrTagNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := tagmock.NewMockRepository(ctrl)
	tagID := uuid.New()

	repo.EXPECT().FindByID(gomock.Any(), tagID).Return(newStoredTag(t, tagID, uuid.New()), nil)
	// Update must not be called for a tag owned by someone else.

	cmd := taguc.NewUpdateTagCommand(repo, markingTxManager{})
	_, _, err := cmd.Execute(context.Background(), uuid.New(), taguc.UpdateTagInput{
		ID: tagID, Name: validName, Color: validColor,
	})

	if !errors.Is(err, domaintag.ErrTagNotFound) {
		t.Errorf("got %v, want ErrTagNotFound", err)
	}
}

func TestUpdateTagCommand_Execute_UpdateReturnsDomainError_PropagatesIt(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{name: "name taken by another tag", err: domaintag.ErrTagNameTaken},
		{name: "tag deleted after it was read", err: domaintag.ErrTagNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := tagmock.NewMockRepository(ctrl)
			userID := uuid.New()
			tagID := uuid.New()

			repo.EXPECT().FindByID(gomock.Any(), tagID).Return(newStoredTag(t, tagID, userID), nil)
			repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(tc.err)

			cmd := taguc.NewUpdateTagCommand(repo, markingTxManager{})
			_, _, err := cmd.Execute(context.Background(), userID, taguc.UpdateTagInput{
				ID: tagID, Name: validName, Color: validColor,
			})

			if !errors.Is(err, tc.err) {
				t.Errorf("got %v, want %v", err, tc.err)
			}
		})
	}
}

func TestUpdateTagCommand_Execute_RepositoryErrors(t *testing.T) {
	cases := []struct {
		name   string
		expect func(repo *tagmock.MockRepository, tagID, userID uuid.UUID)
	}{
		{
			name: "FindByID fails",
			expect: func(repo *tagmock.MockRepository, tagID, _ uuid.UUID) {
				repo.EXPECT().FindByID(gomock.Any(), tagID).Return(nil, errDBFailure)
			},
		},
		{
			name: "Update fails",
			expect: func(repo *tagmock.MockRepository, tagID, userID uuid.UUID) {
				repo.EXPECT().FindByID(gomock.Any(), tagID).Return(newStoredTag(t, tagID, userID), nil)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errDBFailure)
			},
		},
		{
			name: "CountEventsByTagID fails",
			expect: func(repo *tagmock.MockRepository, tagID, userID uuid.UUID) {
				repo.EXPECT().FindByID(gomock.Any(), tagID).Return(newStoredTag(t, tagID, userID), nil)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
				repo.EXPECT().CountEventsByTagID(gomock.Any(), tagID).Return(0, errDBFailure)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := tagmock.NewMockRepository(ctrl)
			userID := uuid.New()
			tagID := uuid.New()
			tc.expect(repo, tagID, userID)

			cmd := taguc.NewUpdateTagCommand(repo, markingTxManager{})
			_, _, err := cmd.Execute(context.Background(), userID, taguc.UpdateTagInput{
				ID: tagID, Name: validName, Color: validColor,
			})

			if !errors.Is(err, errDBFailure) {
				t.Errorf("got %v, want wrapped errDBFailure", err)
			}
		})
	}
}
