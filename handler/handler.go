package handler

import (
	"fairytale-creator/flag"

	"github.com/gin-gonic/gin"
)

const (
	Data    = "data"
	Message = "message"
)

func Init(engine *gin.Engine) {
	user := engine.Group("/v1/user")
	{
		user.POST("/login", login)
	}

	story := engine.Group("/v1/story")
	{
		story.POST("/add", addStory)
		story.GET("/list", listStory)
		story.POST("/voice/generate", generateVoice)
	}
	engine.Static("/v1/resource", flag.VideoRoot)
}
