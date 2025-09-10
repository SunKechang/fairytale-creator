package service

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"fairytale-creator/flag"
	"fairytale-creator/logger"
	"fairytale-creator/request"
)

type UserService struct {
}

func NewUserService() UserService {
	return UserService{}
}

func (p *UserService) Login(c *gin.Context, req request.LoginReq) bool {
	logger.Log(req.Username, req.Password)
	if req.Username != flag.Username {
		return false
	}
	if req.Password != flag.Password {
		return false
	}
	session := sessions.Default(c)
	session.Set("login", true)
	err := session.Save()
	if err != nil {
		return false
	}
	return true
}
