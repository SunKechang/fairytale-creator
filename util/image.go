package util

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func SaveImage(url, path string) error {
	// 发送HTTP GET请求获取图片
	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer response.Body.Close()

	// 检查HTTP响应状态码
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed with status: %s", response.Status)
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 创建目标文件
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// 将响应体内容复制到文件
	_, err = io.Copy(file, response.Body)
	if err != nil {
		// 如果复制失败，删除可能已创建的不完整文件
		os.Remove(path)
		return fmt.Errorf("failed to save image: %w", err)
	}

	return nil
}
