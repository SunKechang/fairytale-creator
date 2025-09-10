package modelapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// TTSClient 文本转语音客户端
type CosyVoiceClient struct {
	APIKey     string
	OutputFile string
	Voice      string
	Format     string
	SampleRate int
	Volume     int
	Rate       int
	Pitch      int
	conn       *websocket.Conn
}

// NewCosyVoiceClient 创建新的TTS客户端
func NewCosyVoiceClient(apiKey, outputFile string) *CosyVoiceClient {
	return &CosyVoiceClient{
		APIKey:     apiKey,
		OutputFile: outputFile,
		Voice:      "longyuan_v2", // 默认声音
		Format:     "mp3",         // 默认格式
		SampleRate: 22050,         // 默认采样率
		Volume:     50,            // 默认音量
		Rate:       1,             // 默认语速
		Pitch:      1,             // 默认音调
	}
}

// 设置声音参数
func (c *CosyVoiceClient) SetVoiceParams(voice, format string, sampleRate, volume, rate, pitch int) {
	if voice != "" {
		c.Voice = voice
	}
	if format != "" {
		c.Format = format
	}
	if sampleRate > 0 {
		c.SampleRate = sampleRate
	}
	if volume >= 0 {
		c.Volume = volume
	}
	if rate > 0 {
		c.Rate = rate
	}
	if pitch > 0 {
		c.Pitch = pitch
	}
}

// 合成文本为语音
func (c *CosyVoiceClient) Synthesize(texts []string) error {
	// 检查并清空输出文件
	if err := c.clearOutputFile(); err != nil {
		return fmt.Errorf("清空输出文件失败: %v", err)
	}

	// 连接WebSocket服务
	conn, err := c.connectWebSocket()
	if err != nil {
		return fmt.Errorf("连接WebSocket失败: %v", err)
	}
	c.conn = conn
	defer c.closeConnection()

	// 启动一个goroutine来接收结果
	done, taskStarted := c.startResultReceiver()

	// 发送run-task指令
	taskID, err := c.sendRunTaskCmd()
	if err != nil {
		return fmt.Errorf("发送run-task指令失败: %v", err)
	}

	// 等待task-started事件
	for !*taskStarted {
		time.Sleep(100 * time.Millisecond)
	}

	// 发送待合成文本
	if err := c.sendContinueTaskCmd(taskID, texts); err != nil {
		return fmt.Errorf("发送待合成文本失败: %v", err)
	}

	// 发送finish-task指令
	if err := c.sendFinishTaskCmd(taskID); err != nil {
		return fmt.Errorf("发送finish-task指令失败: %v", err)
	}

	// 等待接收结果的goroutine完成
	<-done

	return nil
}

// 内部结构体和方法的实现与示例代码类似，但做了适当调整以支持客户端模式
// 以下是简化的实现，完整实现需要包含所有必要的方法

const wsURL = "wss://dashscope.aliyuncs.com/api-ws/v1/inference/"

// 定义结构体来表示JSON数据
type Header struct {
	Action       string                 `json:"action"`
	TaskID       string                 `json:"task_id"`
	Streaming    string                 `json:"streaming"`
	Event        string                 `json:"event"`
	ErrorCode    string                 `json:"error_code,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Attributes   map[string]interface{} `json:"attributes"`
}

type Payload struct {
	TaskGroup  string     `json:"task_group"`
	Task       string     `json:"task"`
	Function   string     `json:"function"`
	Model      string     `json:"model"`
	Parameters Params     `json:"parameters"`
	Resources  []Resource `json:"resources"`
	Input      Input      `json:"input"`
}

type Params struct {
	TextType   string `json:"text_type"`
	Voice      string `json:"voice"`
	Format     string `json:"format"`
	SampleRate int    `json:"sample_rate"`
	Volume     int    `json:"volume"`
	Rate       int    `json:"rate"`
	Pitch      int    `json:"pitch"`
}

type Resource struct {
	ResourceID   string `json:"resource_id"`
	ResourceType string `json:"resource_type"`
}

type Input struct {
	Text string `json:"text"`
}

type Event struct {
	Header  Header  `json:"header"`
	Payload Payload `json:"payload"`
}

var dialer = websocket.DefaultDialer

// 连接WebSocket服务
func (c *CosyVoiceClient) connectWebSocket() (*websocket.Conn, error) {
	header := make(http.Header)
	header.Add("X-DashScope-DataInspection", "enable")
	header.Add("Authorization", fmt.Sprintf("bearer %s", c.APIKey))
	conn, _, err := dialer.Dial(wsURL, header)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// 发送run-task指令
func (c *CosyVoiceClient) sendRunTaskCmd() (string, error) {
	taskID := uuid.New().String()
	runTaskCmd := Event{
		Header: Header{
			Action:    "run-task",
			TaskID:    taskID,
			Streaming: "duplex",
		},
		Payload: Payload{
			TaskGroup: "audio",
			Task:      "tts",
			Function:  "SpeechSynthesizer",
			Model:     "cosyvoice-v2",
			Parameters: Params{
				TextType:   "PlainText",
				Voice:      c.Voice,
				Format:     c.Format,
				SampleRate: c.SampleRate,
				Volume:     c.Volume,
				Rate:       c.Rate,
				Pitch:      c.Pitch,
			},
			Input: Input{},
		},
	}
	runTaskCmdJSON, err := json.Marshal(runTaskCmd)
	if err != nil {
		return "", err
	}
	err = c.conn.WriteMessage(websocket.TextMessage, runTaskCmdJSON)
	return taskID, err
}

// 发送待合成文本
func (c *CosyVoiceClient) sendContinueTaskCmd(taskID string, texts []string) error {
	for _, text := range texts {
		runTaskCmd := Event{
			Header: Header{
				Action:    "continue-task",
				TaskID:    taskID,
				Streaming: "duplex",
			},
			Payload: Payload{
				Input: Input{
					Text: text,
				},
			},
		}
		runTaskCmdJSON, err := json.Marshal(runTaskCmd)
		if err != nil {
			return err
		}

		err = c.conn.WriteMessage(websocket.TextMessage, runTaskCmdJSON)
		if err != nil {
			return err
		}
	}

	return nil
}

// 启动一个goroutine来接收结果
func (c *CosyVoiceClient) startResultReceiver() (chan struct{}, *bool) {
	done := make(chan struct{})
	taskStarted := new(bool)
	*taskStarted = false

	go func() {
		defer close(done)
		for {
			msgType, message, err := c.conn.ReadMessage()
			if err != nil {
				fmt.Println("解析服务器消息失败：", err)
				return
			}

			if msgType == websocket.BinaryMessage {
				// 处理二进制音频流
				if err := c.writeBinaryDataToFile(message); err != nil {
					fmt.Println("写入二进制数据失败：", err)
					return
				}
			} else {
				// 处理文本消息
				var event Event
				err = json.Unmarshal(message, &event)
				if err != nil {
					fmt.Println("解析事件失败：", err)
					continue
				}
				if c.handleEvent(event, taskStarted) {
					return
				}
			}
		}
	}()

	return done, taskStarted
}

// 处理事件
func (c *CosyVoiceClient) handleEvent(event Event, taskStarted *bool) bool {
	switch event.Header.Event {
	case "task-started":
		fmt.Println("收到task-started事件")
		*taskStarted = true
	case "result-generated":
		// 忽略result-generated事件
		return false
	case "task-finished":
		fmt.Println("任务完成")
		return true
	case "task-failed":
		c.handleTaskFailed(event)
		return true
	default:
		fmt.Printf("预料之外的事件：%v\n", event)
	}
	return false
}

// 处理任务失败事件
func (c *CosyVoiceClient) handleTaskFailed(event Event) {
	if event.Header.ErrorMessage != "" {
		fmt.Printf("任务失败：%s\n", event.Header.ErrorMessage)
	} else {
		fmt.Println("未知原因导致任务失败")
	}
}

// 关闭连接
func (c *CosyVoiceClient) closeConnection() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// 写入二进制数据到文件
func (c *CosyVoiceClient) writeBinaryDataToFile(data []byte) error {
	file, err := os.OpenFile(c.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

// 发送finish-task指令
func (c *CosyVoiceClient) sendFinishTaskCmd(taskID string) error {
	finishTaskCmd := Event{
		Header: Header{
			Action:    "finish-task",
			TaskID:    taskID,
			Streaming: "duplex",
		},
		Payload: Payload{
			Input: Input{},
		},
	}
	finishTaskCmdJSON, err := json.Marshal(finishTaskCmd)
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(websocket.TextMessage, finishTaskCmdJSON)
}

// 清空输出文件
func (c *CosyVoiceClient) clearOutputFile() error {
	file, err := os.OpenFile(c.OutputFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	return file.Close()
}

// 使用示例
func test() {
	// 从环境变量获取API Key
	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	if apiKey == "" {
		fmt.Println("请设置环境变量 DASHSCOPE_API_KEY")
		return
	}

	// 创建TTS客户端
	client := NewCosyVoiceClient(apiKey, "output.mp3")

	// 可选：设置声音参数
	client.SetVoiceParams("longxiaochun_v2", "mp3", 22050, 50, 1, 1)

	// 要转换的文本
	texts := []string{"床前明月光", "疑是地上霜", "举头望明月", "低头思故乡"}

	// 执行文本转语音
	err := client.Synthesize(texts)
	if err != nil {
		fmt.Printf("语音合成失败: %v\n", err)
		return
	}

	fmt.Println("语音合成完成，输出文件:", client.OutputFile)
}
