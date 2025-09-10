package util

import (
	"fmt"
	"os/exec"
)

// ConvertVideoToMP4 使用FFmpeg将视频文件转换为MP4格式
func ConvertVideoToMP4(inputFile, outputFile string) error {
	// 构建FFmpeg命令
	args := []string{
		"-i", inputFile, // 输入文件
		"-c:v", "libx264", // 视频编码器
		"-c:a", "aac", // 音频编码器
		"-strict", "experimental",
		"-pix_fmt", "yuv420p", // pixel format
		outputFile, // 输出文件
	}

	// 执行FFmpeg命令
	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		return err
	}

	return nil
}
