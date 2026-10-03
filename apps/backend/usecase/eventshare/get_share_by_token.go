package eventshare

import (
	"context"
	"errors"
	"time"

	domaineventshare "github.com/Haya372/ai-trial/backend/domain/eventshare"
	"github.com/Haya372/ai-trial/backend/domain/eventsubscription"
	"github.com/Haya372/ai-trial/backend/domain/user"
)

type GetShareByTokenInput struct {
	Token  domaineventshare.Token
	Viewer user.User
}

type GetShareByTokenOutput struct {
	Title        string
	Description  string
	StartAt      time.Time
	EndAt        time.Time
	Location     string
	URL          string
	IsOwnEvent   bool
	IsSubscribed bool
}

type GetShareByTokenQuery struct {
	loader   *ShareTokenLoader
	subsRepo eventsubscription.Repository
}

func NewGetShareByTokenQuery(
	l *ShareTokenLoader,
	s eventsubscription.Repository,
) *GetShareByTokenQuery {
	return &GetShareByTokenQuery{loader: l, subsRepo: s}
}

func (q *GetShareByTokenQuery) Execute(
	ctx context.Context,
	in GetShareByTokenInput,
) (GetShareByTokenOutput, error) {
	_, ev, err := q.loader.Load(ctx, in.Token)
	if err != nil {
		return GetShareByTokenOutput{}, err
	}

	out := GetShareByTokenOutput{
		Title:       ev.Title,
		Description: ev.Description,
		StartAt:     ev.StartAt,
		EndAt:       ev.EndAt,
		Location:    ev.Location,
		URL:         ev.URL,
	}

	if in.Viewer == nil {
		return out, nil
	}

	out.IsOwnEvent = ev.UserID == in.Viewer.ID()

	_, err = q.subsRepo.FindByEventAndUserID(ctx, ev.ID, in.Viewer.ID())
	if err == nil {
		out.IsSubscribed = true
		return out, nil
	}
	if errors.Is(err, eventsubscription.ErrEventSubscriptionNotFound) {
		return out, nil
	}
	return GetShareByTokenOutput{}, err
}
