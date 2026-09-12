package handler

import (
	"github.com/gin-gonic/gin"
	"hzycoder.com/lion/internal/middleware"
	resDto "hzycoder.com/lion/internal/model/dto/response"
	"hzycoder.com/lion/internal/service"
	"hzycoder.com/lion/pkg/response"
)

type UserHandler struct {
	users *service.UserService
}

func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) GetByUsername(c *gin.Context) {
	user, err := h.users.GetByUsername(c.Request.Context(), c.Param("username"))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, &resDto.QueryUser{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Role:     user.Role,
	})
}

func (h *UserHandler) Me(c *gin.Context) {
	user, err := h.users.GetByID(c.Request.Context(), middleware.GetUserID(c))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, &resDto.QueryUser{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Role:     user.Role,
	})
}
