package eventshare

import (
	"context"
	"fmt"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	domaineventshare "github.com/Haya372/ai-trial/backend/domain/eventshare"
)

type ShareTokenLoader struct {
	shareRepo domaineventshare.Repository
	eventRepo domainevent.Repository
}

func NewShareTokenLoader(
	sr domaineventshare.Repository,
	er domainevent.Repository,
) *ShareTokenLoader {
	return &ShareTokenLoader{shareRepo: sr, eventRepo: er}
}

func (l *ShareTokenLoader) Load(
	ctx context.Context,
	token string,
) (domaineventshare.EventShare, domainevent.Event, error) {
	share, err := l.shareRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, nil, err
	}
	if share.IsExpired() {
		return nil, nil, domaineventshare.ErrEventShareExpired
	}
	ev, err := l.eventRepo.FindByID(ctx, share.EventID())
	if err != nil {
		return nil, nil, fmt.Errorf("load event for share: %w", err)
	}
	return share, ev, nil
}
