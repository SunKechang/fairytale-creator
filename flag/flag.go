package flag

import (
	"flag"
)

var (
	Username              string
	Password              string
	VideoRoot             string
	DeepSeekAPIKey        string
	DeepSeekUrl           string
	JimengAccessKeyID     string
	JimengSecretAccessKey string
	StoryRoot             string
	ImageRoot             string
	CosyVoiceAPIKey       string
	VoiceRoot             string
	DoubaoSeedreamAPIKey  string
)

func init() {
	flag.StringVar(&Username, "username", "admin", "用户名")
	flag.StringVar(&Password, "password", "20240316", "密码")
	flag.StringVar(&VideoRoot, "video-root", "", "视频存储根路径")
	flag.StringVar(&DeepSeekAPIKey, "deepseek-api-key", "", "DeepSeek API Key")
	flag.StringVar(&DeepSeekUrl, "deepseek-url", "https://api.deepseek.com", "DeepSeek URL")
	flag.StringVar(&JimengAccessKeyID, "jimeng-access-key-id", "", "Jimeng Access Key ID")
	flag.StringVar(&JimengSecretAccessKey, "jimeng-secret-access-key", "", "Jimeng Secret Access Key")
	flag.StringVar(&StoryRoot, "story-root", "stories", "故事存储根路径")
	flag.StringVar(&ImageRoot, "image-root", "images", "图片存储根路径")
	flag.StringVar(&CosyVoiceAPIKey, "cosy-voice-api-key", "", "CosyVoice API Key")
	flag.StringVar(&VoiceRoot, "voice-root", "voices", "语音存储根路径")
	flag.StringVar(&DoubaoSeedreamAPIKey, "doubao-seedream-api-key", "", "Doubao Seedream API Key")
	flag.Parse()
}
