package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/session"
)

// issueSession creates a session for userID and returns the plaintext token to
// hand to the client. Only the token's hash is passed to the repository (ADR-033).
func issueSession(
	ctx context.Context, repo session.Repository, userID uuid.UUID, expiresAt time.Time,
) (session.Token, error) {
	token, err := session.GenerateToken()
	if err != nil {
		return session.Token{}, fmt.Errorf("generate session token: %w", err)
	}
	if _, err := repo.Create(ctx, userID, token.Hash(), expiresAt); err != nil {
		return session.Token{}, fmt.Errorf("create session: %w", err)
	}
	return token, nil
}
