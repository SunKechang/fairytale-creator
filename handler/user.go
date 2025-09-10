package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"fairytale-creator/request"
	"fairytale-creator/service"
)

func login(ctx *gin.Context) {
	res := gin.H{
		Data:    nil,
		Message: "",
	}
	defer func() {
		ctx.JSON(http.StatusOK, res)
	}()
	//转化为LoginRequest结构
	var form request.LoginReq
	err := ctx.ShouldBindJSON(&form)
	if err != nil {
		res[Message] = "请求有误"
		return
	}
	userService := service.NewUserService()
	data := userService.Login(ctx, form)
	res[Data] = data
}
