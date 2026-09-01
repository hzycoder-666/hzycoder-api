package service

import (
	"context"
	"database/sql"
	"errors"

	"hzycoder.com/go-gin-template/internal/model"
	"hzycoder.com/go-gin-template/internal/repository"
	"hzycoder.com/go-gin-template/pkg/response"
)

type UserService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	user, err := s.users.FindByUsername(ctx, username)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, response.NewBizError(response.CodeUserNotFound)
	}
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}

	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, userID int64) (*model.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, response.NewBizError(response.CodeUserNotFound)
	}
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}

	return user, nil
}
