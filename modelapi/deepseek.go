package modelapi

import (
	"bytes"
	"encoding/json"
	"fairytale-creator/logger"
	"fairytale-creator/response"
	"fairytale-creator/util"
	"fmt"
	"io/ioutil"
	"net/http"
)

type DeepSeekClient struct {
	APIKey     string
	BaseURL    string
	HttpClient *http.Client
}

func NewDeepSeekClient(apiKey string, baseURL string) *DeepSeekClient {
	return &DeepSeekClient{
		APIKey:     apiKey,
		BaseURL:    baseURL,
		HttpClient: &http.Client{},
	}
}

func (c *DeepSeekClient) GenerateFairyTale(theme string, date string, style []string) (*response.Story, error) {
	prompt := fmt.Sprintf(`请创作一个关于"%s"的童话故事，供儿童阅读。要求：
    1. 故事包含3-4个章节，总体字数在2000字左右
    2. 每个章节需要提供详细的图片描述，用于AI绘画
    3. 为整个故事推荐一个背景音乐风格
    4. 今天是%s，请确保故事内容新颖不重复，但故事中不要出现明确的时间信息
    5. 图片描述提示词以如下方式进行拼接：[ 主体描述 ] + [ 风格设定 ] + [ 细节要求 ] + [ 视觉氛围 ] + [ 图像宽高像素值 ]。其中宽高像素值固定为 2560x1440。参考示例：
        "新中式动漫插画，一个穿着绿色恐龙连体睡衣的小男孩，在雨后初晴的阳台上，好奇地伸出手去接屋檐滴落的水珠。天空湛蓝如洗，远处有彩虹和被雨水浸润后更显葱郁的城市楼宇。光线透过云层形成美丽的丁达尔光束，空气中仿佛还弥漫着潮湿清新的味道。"
    6. 图片描述中，风格根据故事内容从以下风格中选出：%s
    请以JSON格式返回，包含以下字段：
    - title: 故事标题
    - author: 作者(可以虚构)
    - description: 故事简介
    - music_style: 背景音乐风格描述
    - chapters: 章节数组，每个章节包含title, content, image_prompt, chapter_number`, theme, date, style)

	requestBody, _ := json.Marshal(map[string]interface{}{
		"model":       "deepseek-chat",
		"messages":    []map[string]string{{"role": "user", "content": prompt}},
		"temperature": 0.8,
		"max_tokens":  8000,
	})

	req, _ := http.NewRequest("POST", c.BaseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	// 解析DeepSeek返回的JSON，提取故事内容
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	content := result["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)

	var story response.Story
	// 清洗掉content中的```json和```
	jsonStr, err := util.ExtractJSON(content)
	if err != nil {
		logger.Error("deepseek generate fairy tale - extract json error:", err.Error())
		return nil, err
	}
	err = json.Unmarshal([]byte(jsonStr), &story)
	if err != nil {
		logger.Error("deepseek generate fairy tale - unmarshal json error:", err.Error())
		return nil, err
	}
	story.CreatedAt = date

	return &story, nil
}
