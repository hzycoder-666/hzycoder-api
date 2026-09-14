package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"hzycoder.com/lion/internal/auth"
	"hzycoder.com/lion/internal/model"
	reqDto "hzycoder.com/lion/internal/model/dto/request"
	resDto "hzycoder.com/lion/internal/model/dto/response"
	"hzycoder.com/lion/internal/repository"
	"hzycoder.com/lion/pkg/response"
)

type AuthService struct {
	users              repository.UserRepository
	jwtSecret          string
	jwtExpire          time.Duration
	allowAdminRegister bool
}

type LoginResp struct {
	Token    string            `json:"token"`
	UserInfo *resDto.QueryUser `json:"userInfo"`
}

// dummyPasswordHash 用于用户不存在时的假哈希比较，避免通过响应时间探测用户名是否存在
var dummyPasswordHash = mustBcryptHash("dummy-password")

func mustBcryptHash(password string) []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return hash
}

func NewAuthService(users repository.UserRepository, jwtSecret string, jwtExpire time.Duration, allowAdminRegister bool) *AuthService {
	return &AuthService{
		users:              users,
		jwtSecret:          jwtSecret,
		jwtExpire:          jwtExpire,
		allowAdminRegister: allowAdminRegister,
	}
}

func (s *AuthService) Login(ctx context.Context, req reqDto.LoginUser) (*LoginResp, error) {
	user, err := s.users.FindByUsername(ctx, req.Username)
	if err != nil {
		_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(req.Password))
		return nil, response.NewBizError(response.CodePasswordWrong, "用户名或密码错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, response.NewBizError(response.CodePasswordWrong, "用户名或密码错误")
	}

	token, err := auth.GenerateToken([]byte(s.jwtSecret), s.jwtExpire, user.ID, user.Username, user.Role)
	if err != nil {
		slog.Error("generate token failed", "error", err)
		return nil, response.NewBizError(response.CodeSystemError, "login failed")
	}

	return &LoginResp{
		Token: token,
		UserInfo: &resDto.QueryUser{
			ID:       user.ID,
			Username: user.Username,
			Nickname: user.Nickname,
			Role:     user.Role,
		},
	}, nil
}

func (s *AuthService) Register(ctx context.Context, req reqDto.RegisterUser) (*LoginResp, error) {
	exists, err := s.users.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}
	if exists {
		return nil, response.NewBizError(response.CodeUserExists)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("hash password failed", "error", err)
		return nil, response.NewBizError(response.CodeSystemError)
	}

	role := model.RoleMember
	if s.allowAdminRegister && req.Role != nil && model.IsValid(*req.Role) {
		role = *req.Role
	}

	newUser, err := s.users.Create(ctx, model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Role:     role,
	})
	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, response.NewBizError(response.CodeUserExists)
		}
		return nil, response.NewBizError(response.CodeDBError)
	}

	token, err := auth.GenerateToken([]byte(s.jwtSecret), s.jwtExpire, newUser.ID, newUser.Username, newUser.Role)
	if err != nil {
		slog.Error("generate token failed", "error", err)
		return nil, response.NewBizError(response.CodeSystemError, "register succeeded, please login manually")
	}

	return &LoginResp{
		Token: token,
		UserInfo: &resDto.QueryUser{
			ID:       newUser.ID,
			Username: newUser.Username,
			Nickname: newUser.Nickname,
			Role:     newUser.Role,
		},
	}, nil
}

func isDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 || strings.Contains(err.Error(), "Duplicate")
}
