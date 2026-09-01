package handler

import (
	"github.com/gin-gonic/gin"
	"hzycoder.com/go-gin-template/internal/service"
	"hzycoder.com/go-gin-template/pkg/response"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

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
