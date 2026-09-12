package handler

import (
	"hzycoder.com/lion/internal/model"
	reqDto "hzycoder.com/lion/internal/model/dto/request"
)

type registerRequest struct {
	Username        string      `json:"username" binding:"required,min=3,max=20,alphanum"`
	Password        string      `json:"password" binding:"required,min=8,max=20,password_complexity"`
	ConfirmPassword string      `json:"confirm_password" binding:"required,eqfield=Password"`
	Role            *model.Role `json:"role"`
}

func (r registerRequest) toDTO() reqDto.RegisterUser {
	return reqDto.RegisterUser{
		Username:        r.Username,
		Password:        r.Password,
		ConfirmPassword: r.ConfirmPassword,
		Role:            r.Role,
	}
}

type loginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20,alphanum"`
	Password string `json:"password" binding:"required,min=8,max=20,password_complexity"`
}

func (r loginRequest) toDTO() reqDto.LoginUser {
	return reqDto.LoginUser{
		Username: r.Username,
		Password: r.Password,
	}
}
