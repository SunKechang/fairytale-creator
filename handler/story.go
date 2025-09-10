package handler

import (
	"fairytale-creator/flag"
	"fairytale-creator/request"
	"fairytale-creator/service"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
)

func addStory(c *gin.Context) {
	res := gin.H{
		Data:    nil,
		Message: "",
	}
	defer func() {
		c.JSON(http.StatusOK, res)
	}()
	storyService := service.NewStoryService()
	success := storyService.AddStory()
	if !success {
		res[Message] = "生成故事失败"
		return
	}
	res[Data] = true
	res[Message] = "生成故事成功"
	return

}

func listStory(c *gin.Context) {
	res := gin.H{
		Data:    nil,
		Message: "",
	}
	defer func() {
		c.JSON(http.StatusOK, res)
	}()
	res[Data] = flag.DeepSeekAPIKey
	res[Message] = "获取故事成功"
	return
}

func generateVoice(c *gin.Context) {
	res := gin.H{
		Data:    nil,
		Message: "",
	}
	defer func() {
		c.JSON(http.StatusOK, res)
	}()
	var form request.GenerateVoiceReq
	err := c.ShouldBindJSON(&form)
	if err != nil {
		res[Message] = "请求有误"
		return
	}
	storyService := service.NewStoryService()
	success := storyService.GenerateVoice(form.Text, path.Join(flag.VideoRoot, form.Filename))
	if !success {
		res[Message] = "生成语音失败"
		return
	}
	res[Data] = true
	res[Message] = "生成语音成功"
	return
}
