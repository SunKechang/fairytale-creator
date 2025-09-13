package modelapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"fairytale-creator/logger"
)

// D1QueryRequest 表示发送到 Cloudflare D1 API 的请求结构
type D1QueryRequest struct {
	SQL    string        `json:"sql"`
	Params []interface{} `json:"params,omitempty"`
}

// D1QueryResponse 表示从 Cloudflare D1 API 返回的响应结构
type D1QueryResponse struct {
	Errors   []D1Message `json:"errors,omitempty"`
	Messages []D1Message `json:"messages,omitempty"`
	Result   []D1Result  `json:"result,omitempty"`
	Success  bool        `json:"success"`
}

// D1Message 表示 API 返回的错误或消息
type D1Message struct {
	Code             int    `json:"code"`
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url,omitempty"`
	Source           struct {
		Pointer string `json:"pointer,omitempty"`
	} `json:"source,omitempty"`
}

// D1Result 表示查询结果
type D1Result struct {
	Meta struct {
		ChangedDB       bool    `json:"changed_db"`
		Changes         int     `json:"changes"`
		Duration        float64 `json:"duration"`
		LastRowID       int     `json:"last_row_id"`
		RowsRead        int     `json:"rows_read"`
		RowsWritten     int     `json:"rows_written"`
		ServedByPrimary bool    `json:"served_by_primary"`
		ServedByRegion  string  `json:"served_by_region"`
		SizeAfter       int     `json:"size_after"`
		Timings         struct {
			SQLDurationMS float64 `json:"sql_duration_ms"`
		} `json:"timings"`
	} `json:"meta"`
	Results []map[string]interface{} `json:"results"`
	Success bool                     `json:"success"`
}

// D1Client 表示 Cloudflare D1 客户端
type D1Client struct {
	AccountID  string
	DatabaseID string
	APIKey     string
	HTTPClient *http.Client
}

// NewD1Client 创建一个新的 D1 客户端
func NewD1Client(accountID, databaseID, apiKey string) *D1Client {
	return &D1Client{
		AccountID:  accountID,
		DatabaseID: databaseID,
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}
}

// ExecuteQuery 执行 SQL 查询并返回结果
func (c *D1Client) ExecuteQuery(sql string, params []interface{}) (*D1QueryResponse, error) {
	// 构建请求体
	requestBody := D1QueryRequest{
		SQL:    sql,
		Params: params,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		logger.Error("failed to marshal request body: " + err.Error())
		return nil, err
	}

	// 构建请求 URL
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/d1/database/%s/query",
		c.AccountID, c.DatabaseID)

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("failed to create request: " + err.Error())
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	// 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		logger.Error("failed to send request: " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read response body: " + err.Error())
		return nil, err
	}

	// 解析响应
	var response D1QueryResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		logger.Error("failed to unmarshal response: " + err.Error())
		return nil, err
	}

	// 检查请求是否成功
	if !response.Success {
		if len(response.Errors) > 0 {
			logger.Error("API error: " + response.Errors[0].Message + " (code: " + strconv.Itoa(response.Errors[0].Code) + ")")
			return nil, err
		}
		logger.Error("unknown API error")
		return nil, err
	}

	return &response, nil
}
