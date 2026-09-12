package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/user"
	query "github.com/Haya372/ai-trial/backend/infrastructure/db/generated"
)

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) user.Repository {
	return &userRepository{pool: pool}
}

func (r *userRepository) Create(
	ctx context.Context, email user.Email, displayName string, password user.Password,
) (user.User, error) {
	q := query.New(r.pool)
	row, err := q.InsertUser(ctx, query.InsertUserParams{
		Email:        string(email),
		DisplayName:  displayName,
		PasswordHash: password.Hash(),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, user.ErrEmailTaken
		}
		return nil, err
	}
	return user.New(uuid.UUID(row.ID.Bytes), email, row.DisplayName, password.Hash()), nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email user.Email) (user.User, error) {
	q := query.New(r.pool)
	row, err := q.FindUserByEmail(ctx, string(email))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, user.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user.New(uuid.UUID(row.ID.Bytes), email, row.DisplayName, row.PasswordHash), nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (user.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, email, display_name, password_hash FROM users WHERE id = $1`,
		pgtype.UUID{Bytes: id, Valid: true},
	)
	var (
		pgID         pgtype.UUID
		emailStr     string
		displayName  string
		passwordHash string
	)
	if err := row.Scan(&pgID, &emailStr, &displayName, &passwordHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	email, err := user.NewEmail(emailStr)
	if err != nil {
		return nil, err
	}
	return user.New(uuid.UUID(pgID.Bytes), email, displayName, passwordHash), nil
}
