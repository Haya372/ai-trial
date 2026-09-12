package db_test

import (
	"context"
	"testing"

	"github.com/Haya372/ai-trial/backend/infrastructure/db"
)

func TestGetTx_withoutTx_returnsFalse(t *testing.T) {
	_, ok := db.GetTx(context.Background())
	if ok {
		t.Fatal("expected false when no tx in context")
	}
}
