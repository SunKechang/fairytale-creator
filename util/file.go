package util

import (
	"fmt"
	"os"
)

func DeleteFileIfExist(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，无需删除，返回nil
		}
		// 其他错误
		return err
	}

	// 文件存在，删除文件
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}
	return nil
}
