package util

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

// GenerateRandomString 生成一个指定长度的随机字符串
func GenerateRandomString(length int) (string, error) {
	// 确保length是偶数，因为每个字节将被编码为两个十六进制字符
	if length%2 != 0 {
		return "", fmt.Errorf("length must be even")
	}

	// 创建一个字节切片来存储随机生成的字节
	bytes := make([]byte, length/2)

	// 使用crypto/rand包生成随机字节
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// 将字节切片编码为十六进制字符串
	return hex.EncodeToString(bytes), nil
}

// 提取 JSON 的通用函数
func ExtractJSON(s string) (string, error) {
	// 查找第一个 '{' 的位置
	start := strings.Index(s, "{")
	if start == -1 {
		return "", fmt.Errorf("未找到 JSON 起始标记 '{'")
	}

	// 查找最后一个 '}' 的位置
	end := strings.LastIndex(s, "}")
	if end == -1 {
		return "", fmt.Errorf("未找到 JSON 结束标记 '}'")
	}

	// 确保结束位置在开始位置之后
	if end < start {
		return "", fmt.Errorf("JSON 结束标记在开始标记之前")
	}

	// 提取 JSON 部分
	jsonStr := s[start : end+1]
	return jsonStr, nil
}
