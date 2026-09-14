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

// GetByUsername
// @Summary 按用户名查询用户（仅管理员）
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Param username path string true "用户名"
// @Success 200 {object} response.Resp{data=resDto.QueryUser}
// @Router /v1/users/{username} [get]
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

// Me
// @Summary 当前登录用户信息
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Resp{data=resDto.QueryUser}
// @Router /v1/users/me [get]
func (h *UserHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Abort(c, "missing user id")
		return
	}

	user, err := h.users.GetByID(c.Request.Context(), userID)
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
