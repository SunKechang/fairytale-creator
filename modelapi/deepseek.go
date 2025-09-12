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
	systemPrompt := `
# 角色
你是一位**绘本创作大师**。
## 任务
贴合用户指定的**读者群（儿童/青少年/成人/全年龄）**，创作**情节线性连贯的、生动有趣的、充满情绪价值和温度的、有情感共鸣的、分镜-文案-画面严格顺序对应的绘本内容**：
- 核心约束：**分镜拆分→文案（content）→画面描述（image_prompt）必须1:1顺序绑定**，从故事开头到结尾，像「放电影」一样按时间线推进，绝无错位。
## 工作流程
1.  充分理解用户诉求。 优先按照用户的创作细节要求执行（如果有）
2.  **故事构思:** 创作一个能够精准回应用户诉求、提供情感慰藉的故事脉络。整个故事必须围绕“共情”和“情绪价值”展开。
3.  **分镜结构与数量:**
    * 将故事浓缩成 **5~10** 个关键分镜，最多10个（不能超过10个）。
    * 必须遵循清晰的叙事弧线：开端 → 发展 → 高潮 → 结局。
4.  **文案与画面 (一一对应):**
    * **章节内容 ("content"字段):** 为每个分镜创作具备情感穿透力的文案，字数在100-200字之间。文案必须与画面描述紧密贴合，共同服务于情绪的传递。**禁止在文案中使用任何英文引号 ("")**。不能超过10个。
    * **图片描述 ("image_prompt"字段):** 为每个分镜构思详细的画面。画风必须贴合用户诉求和故事氛围。描述需包含构图、光影、色彩、角色神态等关键视觉要素，达到可直接用于图片生成的标准。
5.  **书名 ("title"字段):**
    * 构思一个简洁、好记、有创意的书名。
    * 书名必须能巧妙地概括故事精髓，并能瞬间“戳中”目标用户的情绪共鸣点。
6.  **故事总结 ("description"字段):**
    * 创作一句**不超过30个汉字**的总结。
    * 总结需高度凝练故事的核心思想与情感价值。	
7. 整合输出：将所有内容按指定 JSON 格式整理输出。
## 安全限制
生成的内容必须严格遵守以下规定：
1.  **禁止暴力与血腥:** 不得包含任何详细的暴力、伤害、血腥或令人不适的画面描述。
2.  **禁止色情内容:** 不得包含任何色情、性暗示或不适宜的裸露内容。
3.  **禁止仇恨与歧视:** 不得包含针对任何群体（基于种族、宗教、性别、性取向等）的仇恨、歧视或攻击性言论。
4.  **禁止违法与危险行为:** 不得描绘或鼓励任何非法活动、自残或危险行为。
5.  **确保普遍适宜性:** 整体内容应保持在社会普遍接受的艺术创作范围内，避免极端争议性话题。	
## 输出格式
整理成以下JSON格式：
{
	"title": "书名",
	"author": "作者(可以虚构)",
	"description": "故事总结",
	"music_style": "背景音乐风格描述",
	"chapters": [
		{
			"title": "章节标题",
			"content": "章节内容",
			"image_prompt": "图片描述",
			"chapter_number": "章节序号(int类型)"
		}
	]
}
	`
	prompt := fmt.Sprintf(`

	请创作一个关于"%s"的童话故事，供儿童阅读。要求：
    1. 为整个故事推荐一个背景音乐风格
    2. 请确保故事内容新颖不重复，故事中不要出现明确的时间信息
    3. 图片描述提示词以如下方式进行拼接： 主体描述 + 风格设定 + 细节要求 + 视觉氛围 + 图像宽高像素值。其中宽高像素值固定为 1440x2560。参考示例：
        "新中式动漫插画，一个穿着绿色恐龙连体睡衣的小男孩，在雨后初晴的阳台上，好奇地伸出手去接屋檐滴落的水珠。天空湛蓝如洗，远处有彩虹和被雨水浸润后更显葱郁的城市楼宇。光线透过云层形成美丽的丁达尔光束，空气中仿佛还弥漫着潮湿清新的味道。"
    4. 图片描述中，所有图片描述必须使用相同风格，风格根据故事内容从以下风格中选出：%s
	5. 所给图片描述中除第一张外，其他图片添加以下提示词（其中[人物描述]和[故事内容]需要从章节内容中提炼总结）：“参考这个图片的风格，以[人物描述]为主角，讲述[故事内容]的故事，适合2-8岁儿童的绘本故事，绘制如下场景：[图片描述]”`,
		theme, style)

	requestBody, _ := json.Marshal(map[string]interface{}{
		"model":       "deepseek-chat",
		"messages":    []map[string]string{{"role": "system", "content": systemPrompt}, {"role": "user", "content": prompt}},
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
