package eventshare_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	eventmock "github.com/Haya372/ai-trial/backend/domain/event/generated"
	domaineventshare "github.com/Haya372/ai-trial/backend/domain/eventshare"
	eventsharemock "github.com/Haya372/ai-trial/backend/domain/eventshare/generated"
	eventshareuc "github.com/Haya372/ai-trial/backend/usecase/eventshare"
)

func TestShareTokenLoader_Load_notFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	shareRepo := eventsharemock.NewMockRepository(ctrl)
	eventRepo := eventmock.NewMockRepository(ctrl)
	shareRepo.EXPECT().FindByToken(gomock.Any(), "tok").Return(nil, domaineventshare.ErrEventShareNotFound)

	loader := eventshareuc.NewShareTokenLoader(shareRepo, eventRepo)
	_, _, err := loader.Load(context.Background(), "tok")

	if !errors.Is(err, domaineventshare.ErrEventShareNotFound) {
		t.Errorf("expected ErrEventShareNotFound, got %v", err)
	}
}

func TestShareTokenLoader_Load_expired(t *testing.T) {
	ctrl := gomock.NewController(t)
	shareRepo := eventsharemock.NewMockRepository(ctrl)
	eventRepo := eventmock.NewMockRepository(ctrl)

	eventID := uuid.New()
	expiredShare := newTestShare(eventID, true)
	shareRepo.EXPECT().FindByToken(gomock.Any(), "tok").Return(expiredShare, nil)

	loader := eventshareuc.NewShareTokenLoader(shareRepo, eventRepo)
	_, _, err := loader.Load(context.Background(), "tok")

	if !errors.Is(err, domaineventshare.ErrEventShareExpired) {
		t.Errorf("expected ErrEventShareExpired, got %v", err)
	}
}

func TestShareTokenLoader_Load_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	shareRepo := eventsharemock.NewMockRepository(ctrl)
	eventRepo := eventmock.NewMockRepository(ctrl)

	ownerID := uuid.New()
	ev := newTestEvent(ownerID)
	share := newTestShare(ev.ID(), false)
	shareRepo.EXPECT().FindByToken(gomock.Any(), "tok").Return(share, nil)
	eventRepo.EXPECT().FindByID(gomock.Any(), ev.ID()).Return(ev, nil)

	loader := eventshareuc.NewShareTokenLoader(shareRepo, eventRepo)
	gotShare, gotEvent, err := loader.Load(context.Background(), "tok")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotShare.ID() != share.ID() {
		t.Errorf("share ID mismatch: got %v, want %v", gotShare.ID(), share.ID())
	}
	if gotEvent.ID() != ev.ID() {
		t.Errorf("event ID mismatch: got %v, want %v", gotEvent.ID(), ev.ID())
	}
}
