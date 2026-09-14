package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"hzycoder.com/lion/internal/model"
	reqDto "hzycoder.com/lion/internal/model/dto/request"
	"hzycoder.com/lion/internal/service"
	"hzycoder.com/lion/pkg/response"
)

type fakeUserRepository struct {
	byUsername map[string]*model.User
	byID       map[int64]*model.User
	exists     bool
	createErr  error
	nextID     int64
}

func (f *fakeUserRepository) FindByUsername(_ context.Context, username string) (*model.User, error) {
	user, ok := f.byUsername[username]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return user, nil
}

func (f *fakeUserRepository) FindByID(_ context.Context, userID int64) (*model.User, error) {
	user, ok := f.byID[userID]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return user, nil
}

func (f *fakeUserRepository) FindRoleByID(_ context.Context, userID int64) (model.Role, error) {
	user, ok := f.byID[userID]
	if !ok {
		return "", sql.ErrNoRows
	}
	return user.Role, nil
}

func (f *fakeUserRepository) ExistsByUsername(context.Context, string) (bool, error) {
	return f.exists, nil
}

func (f *fakeUserRepository) Create(_ context.Context, user model.User) (*model.User, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	if f.nextID == 0 {
		f.nextID = 1
	}
	user.ID = f.nextID
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt
	return &user, nil
}

func TestAuthServiceLogin(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("Demo123!"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	repo := &fakeUserRepository{byUsername: map[string]*model.User{
		"demo001": {
			ID:       1,
			Username: "demo001",
			Password: string(hash),
			Role:     model.RoleMember,
		},
	}}
	svc := service.NewAuthService(repo, "test-secret", time.Hour, false)

	resp, err := svc.Login(context.Background(), reqDto.LoginUser{Username: "demo001", Password: "Demo123!"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if resp.Token == "" || resp.UserInfo.Username != "demo001" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestAuthServiceRegisterDuplicateUser(t *testing.T) {
	repo := &fakeUserRepository{exists: true}
	svc := service.NewAuthService(repo, "test-secret", time.Hour, false)

	_, err := svc.Register(context.Background(), reqDto.RegisterUser{
		Username:        "demo001",
		Password:        "Demo123!",
		ConfirmPassword: "Demo123!",
	})
	if err == nil {
		t.Fatal("expected duplicate user error")
	}

	var bizErr *response.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != response.CodeUserExists {
		t.Fatalf("expected CodeUserExists, got %v", err)
	}
}

func TestAuthServiceRegisterIgnoresRequestedRoleWhenSwitchOff(t *testing.T) {
	repo := &fakeUserRepository{}
	svc := service.NewAuthService(repo, "test-secret", time.Hour, false)

	admin := model.RoleAdmin
	resp, err := svc.Register(context.Background(), reqDto.RegisterUser{
		Username:        "demo001",
		Password:        "Demo123!",
		ConfirmPassword: "Demo123!",
		Role:            &admin,
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if resp.UserInfo.Role != model.RoleMember {
		t.Fatalf("expected role member, got %v", resp.UserInfo.Role)
	}
}
