package modelapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"fairytale-creator/logger"
)

// DoubaoSeedreamClient Doubao Seedream 文生图客户端
type DoubaoSeedreamClient struct {
	APIKey     string
	BaseURL    string
	HttpClient *http.Client
}

// ImageGenerationRequest 图像生成请求结构体
type I2IGenerationRequest struct {
	Model                     string  `json:"model"`
	Prompt                    string  `json:"prompt"`
	Image                     *string `json:"image"`
	Size                      string  `json:"size"`
	SequentialImageGeneration string  `json:"sequential_image_generation"`
	Stream                    bool    `json:"stream"`
	ResponseFormat            string  `json:"response_format"`
	Watermark                 bool    `json:"watermark"`
}

type T2IGenerationRequest struct {
	Model                     string `json:"model"`
	Prompt                    string `json:"prompt"`
	Size                      string `json:"size"`
	SequentialImageGeneration string `json:"sequential_image_generation"`
	Stream                    bool   `json:"stream"`
	ResponseFormat            string `json:"response_format"`
	Watermark                 bool   `json:"watermark"`
}

// ImageGenerationResponse 图像生成响应结构体
type ImageGenerationResponse struct {
	Model   string `json:"model"`
	Created int64  `json:"created"`
	Data    []struct {
		URL  string `json:"url"`
		Size string `json:"size"`
	} `json:"data"`
	Usage struct {
		GeneratedImages int `json:"generated_images"`
		OutputTokens    int `json:"output_tokens"`
		TotalTokens     int `json:"total_tokens"`
	} `json:"usage"`
}

// NewDoubaoSeedreamClient 创建新的 Doubao Seedream 客户端
func NewDoubaoSeedreamClient(apiKey string) *DoubaoSeedreamClient {
	return &DoubaoSeedreamClient{
		APIKey:     apiKey,
		BaseURL:    "https://ark.cn-beijing.volces.com/api/v3/images/generations",
		HttpClient: &http.Client{Timeout: 180 * time.Second},
	}
}

// GenerateImage 生成图像（支持可选 image 参数）
func (c *DoubaoSeedreamClient) GenerateImage(prompt string, imageURL *string) (*ImageGenerationResponse, error) {
	// 构建请求体
	var requestBody interface{}
	if imageURL != nil && *imageURL != "" {
		requestBody = I2IGenerationRequest{
			Model:                     "doubao-seedream-4-0-250828",
			Prompt:                    prompt,
			Image:                     imageURL,
			Size:                      "1440x2560",
			SequentialImageGeneration: "disabled",
			Stream:                    false,
			ResponseFormat:            "url",
			Watermark:                 false,
		}
	} else {
		requestBody = T2IGenerationRequest{
			Model:                     "doubao-seedream-4-0-250828",
			Prompt:                    prompt,
			Size:                      "1440x2560",
			SequentialImageGeneration: "disabled",
			Stream:                    false,
			ResponseFormat:            "url",
			Watermark:                 false,
		}
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		logger.Error("failed to marshal request body: " + err.Error())
		return nil, err
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", c.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("failed to create request: " + err.Error())
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	// 发送请求
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		logger.Error("failed to send request: " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("API request failed with status " + strconv.Itoa(resp.StatusCode) + ": " + string(body))
		return nil, err
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read response body: " + err.Error())
		return nil, err
	}
	logger.Log("generate image response", string(body))

	// 解析响应
	var response ImageGenerationResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		logger.Error("failed to unmarshal response: " + err.Error())
		return nil, err
	}

	// 检查是否有生成的图像
	if len(response.Data) == 0 {
		logger.Error("no images generated")
		return nil, err
	}

	return &response, nil
}

// GenerateImageFromPrompt 仅使用提示词生成图像（向后兼容的方法）
func (c *DoubaoSeedreamClient) GenerateImageFromPrompt(prompt string) (*ImageGenerationResponse, error) {
	return c.GenerateImage(prompt, nil)
}

// GenerateImageFromPromptAndImage 使用提示词和参考图像生成图像
func (c *DoubaoSeedreamClient) GenerateImageFromPromptAndImage(prompt string, imageURL string) (*ImageGenerationResponse, error) {
	return c.GenerateImage(prompt, &imageURL)
}

// GenerateImageAndGetURL 生成图像并返回 URL（简化方法）
func (c *DoubaoSeedreamClient) GenerateImageAndGetURL(prompt string, imageURL *string) (string, error) {
	response, err := c.GenerateImage(prompt, imageURL)
	if err != nil {
		return "", err
	}

	if len(response.Data) > 0 {
		return response.Data[0].URL, nil
	}

	logger.Error("no image URL in response")
	return "", err
}

// GenerateImageFromPromptAndGetURL 仅使用提示词生成图像并返回 URL（简化方法）
func (c *DoubaoSeedreamClient) GenerateImageFromPromptAndGetURL(prompt string) (string, error) {
	return c.GenerateImageAndGetURL(prompt, nil)
}

// GenerateImageFromPromptAndImageAndGetURL 使用提示词和参考图像生成图像并返回 URL（简化方法）
func (c *DoubaoSeedreamClient) GenerateImageFromPromptAndImageAndGetURL(prompt string, imageURL string) (string, error) {
	return c.GenerateImageAndGetURL(prompt, &imageURL)
}
