package db_test

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/Haya372/ai-trial/backend/infrastructure/db"
	"github.com/Haya372/ai-trial/backend/infrastructure/db/mock"
)

var (
	errFn = errors.New("fn error")
	errRb = errors.New("rollback error")
)

func TestRunInTx_rollbackFailure_joinsErrors(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockTx := mock.NewMockTx(ctrl)
	mockTx.EXPECT().Rollback(gomock.Any()).Return(errRb)

	mockBeginner := mock.NewMocktxBeginner(ctrl)
	mockBeginner.EXPECT().Begin(gomock.Any()).Return(mockTx, nil)

	mgr := db.NewPgxTxManagerForTest(mockBeginner)

	err := mgr.RunInTx(context.Background(), func(_ context.Context) error {
		return errFn
	})

	if !errors.Is(err, errFn) {
		t.Errorf("expected err to contain errFn, got: %v", err)
	}
	if !errors.Is(err, errRb) {
		t.Errorf("expected err to contain errRb, got: %v", err)
	}
}
