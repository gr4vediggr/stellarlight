package database

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/gr4vediggr/stellarlight/internal/database/queries"
	"github.com/gr4vediggr/stellarlight/internal/users"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserStore struct {
	queries *queries.Queries
}

func NewPostgresUserStore(db *pgxpool.Pool) *PostgresUserStore {
	return &PostgresUserStore{
		queries: queries.New(db),
	}
}

func (store *PostgresUserStore) GetUserByEmail(ctx context.Context, email string) (*users.User, error) {
	result, err := store.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &users.User{
		ID:          result.ID,
		Email:       result.Email,
		DisplayName: result.DisplayName,
		Password:    result.Password,
		CreatedAt:   result.CreatedAt.Time,
		UpdatedAt:   result.UpdatedAt.Time,
	}, nil
}

func (store *PostgresUserStore) GetUserByID(ctx context.Context, id uuid.UUID) (*users.User, error) {
	result, err := store.queries.GetUserByID(ctx, id)
	log.Println("Fetching user by ID:", id, err)
	if err != nil {
		return nil, err
	}
	return &users.User{
		ID:          result.ID,
		Email:       result.Email,
		DisplayName: result.DisplayName,
		Password:    result.Password,
		CreatedAt:   result.CreatedAt.Time,
		UpdatedAt:   result.UpdatedAt.Time,
	}, nil
}

func (store *PostgresUserStore) CreateUser(ctx context.Context, user *users.User) (*users.User, error) {
	result, err := store.queries.CreateUser(ctx, queries.CreateUserParams{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Password:    user.Password,
	})
	if err != nil {
		return nil, err
	}
	return &users.User{
		ID:          result.ID,
		Email:       result.Email,
		DisplayName: result.DisplayName,
		CreatedAt:   result.CreatedAt.Time,
		UpdatedAt:   result.UpdatedAt.Time,
	}, nil
}

func (store *PostgresUserStore) UpdateUser(ctx context.Context, user *users.User) (*users.User, error) {
	result, err := store.queries.UpdateUser(ctx, queries.UpdateUserParams{
		ID:          user.ID,
		DisplayName: user.DisplayName,
		Password:    user.Password,
	})
	if err != nil {
		return nil, err
	}
	return &users.User{
		ID:          result.ID,
		Email:       result.Email,
		DisplayName: result.DisplayName,
		CreatedAt:   result.CreatedAt.Time,
		UpdatedAt:   result.UpdatedAt.Time,
	}, nil
}

func (store *PostgresUserStore) DeleteUser(ctx context.Context, id uuid.UUID) (*users.User, error) {
	result, err := store.queries.DeleteUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return &users.User{
		ID:          result.ID,
		Email:       result.Email,
		DisplayName: result.DisplayName,
		CreatedAt:   result.CreatedAt.Time,
		UpdatedAt:   result.UpdatedAt.Time,
	}, nil
}

func (store *PostgresUserStore) CreateRefreshToken(ctx context.Context, token *users.CreateRefreshTokenParams) (*users.RefreshToken, error) {
	result, err := store.queries.CreateRefreshToken(ctx, queries.CreateRefreshTokenParams{
		ID:        token.ID,
		UserID:    token.UserID,
		Token:     token.Token,
		ExpiresAt: token.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}
	return &users.RefreshToken{
		ID:        result.ID,
		UserID:    result.UserID,
		Token:     result.Token,
		ExpiresAt: result.ExpiresAt,
		Revoked:   result.Revoked,
		CreatedAt: result.CreatedAt,
	}, nil
}

func (store *PostgresUserStore) DeleteExpiredRefreshTokens(ctx context.Context) error {
	return store.queries.DeleteExpiredRefreshTokens(ctx)
}
func (store *PostgresUserStore) GetRefreshToken(ctx context.Context, token string) (*users.RefreshToken, error) {
	result, err := store.queries.GetRefreshToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return &users.RefreshToken{
		ID:        result.ID,
		UserID:    result.UserID,
		Token:     result.Token,
		ExpiresAt: result.ExpiresAt,
		Revoked:   result.Revoked,
		CreatedAt: result.CreatedAt,
	}, nil
}
func (store *PostgresUserStore) RevokeRefreshToken(ctx context.Context, token string) error {
	return store.queries.RevokeRefreshToken(ctx, token)
}
