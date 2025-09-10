package logger

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"runtime"
	"strings"
)

func Log(text ...string) {
	prefix := "%s:%d "
	content := strings.Join(text, " ")
	_, file, line, _ := runtime.Caller(1)
	str := fmt.Sprintf("[LOG] "+prefix+content+"\n", file, line)
	fmt.Fprintf(gin.DefaultWriter, str)
}

func Error(text ...string) {
	prefix := "%s:%d "
	content := strings.Join(text, " ")
	_, file, line, _ := runtime.Caller(1)
	str := fmt.Sprintf("[ERROR] "+prefix+content+"\n", file, line)
	fmt.Fprintf(gin.DefaultWriter, str)
}
