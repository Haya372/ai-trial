package eventshare

import (
	"context"
	"fmt"

	domaineventshare "github.com/Haya372/ai-trial/backend/domain/eventshare"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

type ShareTokenLoader struct {
	shareRepo  domaineventshare.Repository
	eventQuery eventuc.QueryService
}

func NewShareTokenLoader(
	sr domaineventshare.Repository,
	qs eventuc.QueryService,
) *ShareTokenLoader {
	return &ShareTokenLoader{shareRepo: sr, eventQuery: qs}
}

func (l *ShareTokenLoader) Load(
	ctx context.Context,
	token string,
) (domaineventshare.EventShare, eventuc.EventReadModel, error) {
	share, err := l.shareRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, eventuc.EventReadModel{}, err
	}
	if share.IsExpired() {
		return nil, eventuc.EventReadModel{}, domaineventshare.ErrEventShareExpired
	}
	ev, err := l.eventQuery.FindByID(ctx, share.EventID())
	if err != nil {
		return nil, eventuc.EventReadModel{}, fmt.Errorf("load event for share: %w", err)
	}
	return share, ev, nil
}
