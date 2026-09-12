package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"hzycoder.com/lion/internal/model"
)

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByID(ctx context.Context, userID int64) (*model.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	Create(ctx context.Context, user model.User) (*model.User, error)
}

type SQLUserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *SQLUserRepository {
	return &SQLUserRepository{db: db}
}

func (r *SQLUserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user model.User
	err := r.db.GetContext(ctx, &user, `
		SELECT id, username, password, nickname, role, created_at, updated_at
		FROM users
		WHERE username = ?
	`, username)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *SQLUserRepository) FindByID(ctx context.Context, userID int64) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user model.User
	err := r.db.GetContext(ctx, &user, `
		SELECT id, username, password, nickname, role, created_at, updated_at
		FROM users
		WHERE id = ?
	`, userID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *SQLUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var exists int
	err := r.db.GetContext(ctx, &exists, `SELECT 1 FROM users WHERE username = ? LIMIT 1`, username)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check user exists: %w", err)
	}

	return true, nil
}

func (r *SQLUserRepository) Create(ctx context.Context, user model.User) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	result, err := r.db.NamedExecContext(ctx, `
		INSERT INTO users (username, password, nickname, role, created_at, updated_at)
		VALUES (:username, :password, :nickname, :role, :created_at, :updated_at)
	`, user)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted user id: %w", err)
	}
	user.ID = id

	return &user, nil
}
