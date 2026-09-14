package handler

import (
	"github.com/gin-gonic/gin"
	"hzycoder.com/lion/internal/service"
	"hzycoder.com/lion/pkg/response"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Login
// @Summary 用户登录
// @Description 校验用户名密码，签发 JWT
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body loginRequest true "登录参数"
// @Success 200 {object} response.Resp{data=service.LoginResp}
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	loginResp, err := h.auth.Login(c.Request.Context(), req.toDTO())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, loginResp)
}

// Register
// @Summary 用户注册
// @Description 注册新用户，默认角色 member；仅当服务端开启 auth.allow_admin_register 时才接受请求中的 role 字段
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body registerRequest true "注册参数"
// @Success 200 {object} response.Resp{data=service.LoginResp}
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	registerResp, err := h.auth.Register(c.Request.Context(), req.toDTO())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, registerResp)
}
