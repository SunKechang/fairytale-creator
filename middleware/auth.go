package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

func LoginAuth(c *gin.Context) {
	session := sessions.Default(c)
	login := session.Get("login")
	if login == nil || !login.(bool) {
		c.JSON(http.StatusUnauthorized, gin.H{"data": false, "message": "请登录后再试"})
		c.Abort()
		return
	}
}
