package main

import (
	"context"
	"errors"
	"fairytale-creator/database"
	"fairytale-creator/handler"
	"fairytale-creator/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化数据库
	database.Init()

	r := gin.Default()

	// 创建一个基于内存的存储对象
	store := memstore.NewStore([]byte("secret"))
	// 使用 sessions 中间件，并将内存存储对象传递给它
	r.Use(sessions.Sessions("session", store))
	handler.Init(r)
	srv := &http.Server{
		//0.0.0.0:8080
		Addr:    "127.0.0.1:9700",
		Handler: r,
	}
	// 开始监听
	go func() {
		logger.Log("server started on:", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen error:", err.Error())
		}
	}()

	// 优雅关机
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	sig := <-c
	logger.Log("[main],app stopping,receive:", sig.String())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// stop main apps
	// 关闭 HTTP 服务器
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("服务器关闭错误:", err.Error())
	}

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown:", err.Error())
	}
	logger.Log("[main],app stopped,receive:", sig.String())
}
