package modelapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fairytale-creator/logger"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// 客户端结构体
type JimengClient struct {
	AccessKeyID     string
	SecretAccessKey string
	Addr            string
	Path            string
	Service         string
	Region          string
	Action          string
	Version         string
}

// 任务提交请求结构
type SubmitTaskRequest struct {
	ReqKey           string                 `json:"req_key"`
	Prompt           string                 `json:"prompt"`
	Style            string                 `json:"style,omitempty"`
	Width            int                    `json:"width,omitempty"`
	Height           int                    `json:"height,omitempty"`
	ImageNum         int                    `json:"image_num,omitempty"`
	ImageUrls        []string               `json:"image_urls,omitempty"`
	AdditionalParams map[string]interface{} `json:"additional_params,omitempty"`
}

// 任务查询请求结构
type QueryTaskRequest struct {
	ReqKey  string `json:"req_key"`
	TaskID  string `json:"task_id"`
	ReqJson string `json:"req_json"`
}

// 通用响应结构
type ApiResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 任务提交响应结构
type SubmitTaskResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		TaskID string `json:"task_id"`
	} `json:"data"`
}

// 任务查询响应结构
type QueryTaskResponse struct {
	Code    int    `json:"code"`
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Status           string   `json:"status"`
		BinaryDataBase64 string   `json:"binary_data_base64"`
		ImageUrls        []string `json:"image_urls"`
	} `json:"data"`
	TimeElapsed string `json:"time_elapsed"`
}

// 错误消息常量
const (
	ErrorPreImgRiskNotPass   = "预图风险不通过"
	ErrorPostImgRiskNotPass  = "后图风险不通过"
	ErrorTextRiskNotPass     = "文本风险不通过"
	ErrorPostTextRiskNotPass = "后文本风险不通过"
	ErrorAPILimit            = "QPS超限"
	ErrorConcurrentLimit     = "并发超限"
	ErrorInternal            = "内部错误"
	ErrorInternalRPC         = "内部RPC错误"

	SubmitAction = "CVSync2AsyncSubmitTask"
	QueryAction  = "CVSync2AsyncGetResult"
	Version      = "2022-08-31"
)

// 新建客户端
func NewJimengClient(accessKeyID, secretAccessKey string) *JimengClient {
	return &JimengClient{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		Addr:            "https://visual.volcengineapi.com",
		Path:            "/",
		Service:         "cv",
		Region:          "cn-north-1",
	}
}

// HMAC-SHA256加密
func hmacSHA256(key []byte, content string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(content))
	return mac.Sum(nil)
}

// 获取签名密钥
func getSignedKey(secretKey, date, region, service string) []byte {
	kDate := hmacSHA256([]byte(secretKey), date)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	kSigning := hmacSHA256(kService, "request")
	return kSigning
}

// SHA256哈希
func hashSHA256(data []byte) []byte {
	hash := sha256.New()
	if _, err := hash.Write(data); err != nil {
		log.Printf("input hash err:%s", err.Error())
	}
	return hash.Sum(nil)
}

// 执行请求
func (c *JimengClient) doRequest(method string, action string, version string, queries url.Values, body []byte) ([]byte, int, error) {
	logger.Log("requestBody:", string(body))
	// 构建请求URL
	queries.Set("Action", action)
	queries.Set("Version", version)
	requestAddr := fmt.Sprintf("%s%s?%s", c.Addr, c.Path, queries.Encode())

	request, err := http.NewRequest(method, requestAddr, bytes.NewBuffer(body))
	if err != nil {
		return nil, 0, fmt.Errorf("bad request: %w", err)
	}

	// 构建签名材料
	now := time.Now()
	date := now.UTC().Format("20060102T150405Z")
	authDate := date[:8]
	request.Header.Set("X-Date", date)

	payload := hex.EncodeToString(hashSHA256(body))
	request.Header.Set("X-Content-Sha256", payload)
	request.Header.Set("Content-Type", "application/json")

	queryString := strings.Replace(queries.Encode(), "+", "%20", -1)
	signedHeaders := []string{"host", "x-date", "x-content-sha256", "content-type"}
	var headerList []string
	for _, header := range signedHeaders {
		if header == "host" {
			headerList = append(headerList, header+":"+request.Host)
		} else {
			v := request.Header.Get(header)
			headerList = append(headerList, header+":"+strings.TrimSpace(v))
		}
	}
	headerString := strings.Join(headerList, "\n")

	canonicalString := strings.Join([]string{
		method,
		c.Path,
		queryString,
		headerString + "\n",
		strings.Join(signedHeaders, ";"),
		payload,
	}, "\n")

	hashedCanonicalString := hex.EncodeToString(hashSHA256([]byte(canonicalString)))

	credentialScope := authDate + "/" + c.Region + "/" + c.Service + "/request"
	signString := strings.Join([]string{
		"HMAC-SHA256",
		date,
		credentialScope,
		hashedCanonicalString,
	}, "\n")

	// 构建认证请求头
	signedKey := getSignedKey(c.SecretAccessKey, authDate, c.Region, c.Service)
	signature := hex.EncodeToString(hmacSHA256(signedKey, signString))

	authorization := "HMAC-SHA256" +
		" Credential=" + c.AccessKeyID + "/" + credentialScope +
		", SignedHeaders=" + strings.Join(signedHeaders, ";") +
		", Signature=" + signature

	request.Header.Set("Authorization", authorization)

	// 发起请求
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, 0, fmt.Errorf("do request err: %w", err)
	}
	defer response.Body.Close()

	// 读取响应
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("read response body err: %w", err)
	}

	logger.Log("doRequest", method, "response status", response.Status, "response body", string(bodyBytes))

	return bodyBytes, response.StatusCode, nil
}

// 提交文生图、图生图任务
func (c *JimengClient) SubmitTask(prompt string, options map[string]interface{},
	action string, version string, reqKey string) (string, error) {
	// 构建请求体
	reqBody := SubmitTaskRequest{
		ReqKey: reqKey,
		Prompt: prompt,
	}

	// 设置可选参数
	if style, ok := options["style"].(string); ok {
		reqBody.Style = style
	}
	if width, ok := options["width"].(int); ok {
		reqBody.Width = width
	}
	if height, ok := options["height"].(int); ok {
		reqBody.Height = height
	}
	if imageNum, ok := options["image_num"].(int); ok {
		reqBody.ImageNum = imageNum
	}
	if imageUrls, ok := options["image_urls"].([]string); ok {
		if len(imageUrls) == 1 {
			reqBody.ImageUrls = imageUrls
		}
	}

	reqBodyStr, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request body error: %w", err)
	}

	// 发送请求
	responseBody, statusCode, err := c.doRequest("POST", SubmitAction, Version, url.Values{}, reqBodyStr)
	if err != nil {
		return "", fmt.Errorf("request error: %w", err)
	}

	// 检查HTTP状态码
	if statusCode != 200 {
		return "", fmt.Errorf("API returned HTTP %d", statusCode)
	}

	// 解析响应
	var response SubmitTaskResponse
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return "", fmt.Errorf("unmarshal response error: %w", err)
	}

	// 检查API状态码
	if response.Status != 10000 { // 10000表示成功
		return "", fmt.Errorf("API error: %s", response.Message)
	}

	return response.Data.TaskID, nil
}

func (c *JimengClient) QueryTaskInCircle(reqKey string, taskID string) (string, error) {
	startTime := time.Now()
	time.Sleep(2 * time.Second)
	logger.Log("QueryTaskInCircle", reqKey, taskID, "starttime", startTime.Format("2006-01-02 15:04:05"))
	for {
		queryTaskResponse, err := c.QueryTask(reqKey, taskID)
		if err != nil {
			return "", err
		}
		if queryTaskResponse.Code == 10000 && queryTaskResponse.Message == "Success" && len(queryTaskResponse.Data.ImageUrls) > 0 {
			logger.Log("QueryTaskInCircle", reqKey, taskID, "remaintime", time.Since(startTime).String())
			return queryTaskResponse.Data.ImageUrls[0], nil
		}
		time.Sleep(2 * time.Second)
	}
}

// 查询任务状态
func (c *JimengClient) QueryTask(reqKey string, taskID string) (*QueryTaskResponse, error) {
	// 构建请求体
	reqBody := QueryTaskRequest{
		ReqKey:  reqKey,
		TaskID:  taskID,
		ReqJson: "{\"logo_info\":{\"add_logo\":false,\"position\":0,\"language\":0,\"opacity\":0.3,\"logo_text_content\":\"这里是明水印内容\"},\"return_url\":true,\"aigc_meta\":{\"content_producer\":\"xxx\",\"producer_id\":\"xxx\",\"logo_text_content\":\"xxx\",\"logo_text_content\":\"xxx\"}}",
	}

	reqBodyStr, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body error: %w", err)
	}

	// 发送请求
	responseBody, statusCode, err := c.doRequest("POST", QueryAction, Version, url.Values{}, reqBodyStr)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}

	// 检查HTTP状态码
	if statusCode != 200 {
		return nil, fmt.Errorf("API returned HTTP %d", statusCode)
	}

	// 解析响应
	var response QueryTaskResponse
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response error: %w", err)
	}

	// 检查API状态码
	if response.Status != 10000 { // 10000表示成功
		return nil, fmt.Errorf("API error: %s", response.Message)
	}

	return &response, nil
}
