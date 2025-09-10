package service

import (
	"encoding/json"
	"errors"
	"fairytale-creator/flag"
	"fairytale-creator/logger"
	"fairytale-creator/modelapi"
	"fairytale-creator/util"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	JimengVersion   = "2022-08-31"
	JimengAction    = "CVSync2AsyncSubmitTask"
	JimengReqKeyT2i = "jimeng_t2i_v31" // 文生图
	JimengReqKeyI2i = "jimeng_i2i_v30" // 图生图
)

type StoryService struct {
	DeepSeekAPIKey        string
	DeepSeekUrl           string
	JimengAccessKeyID     string
	JimengSecretAccessKey string
}

func NewStoryService() *StoryService {
	return &StoryService{
		DeepSeekAPIKey:        flag.DeepSeekAPIKey,
		DeepSeekUrl:           flag.DeepSeekUrl,
		JimengAccessKeyID:     flag.JimengAccessKeyID,
		JimengSecretAccessKey: flag.JimengSecretAccessKey,
	}
}

func (s *StoryService) AddStory() bool {
	currentDate := time.Now().Format("2006-01-02")
	theme := util.GenerateDailyTheme(currentDate)
	client := modelapi.NewDeepSeekClient(s.DeepSeekAPIKey, s.DeepSeekUrl)
	story, err := client.GenerateFairyTale(theme, currentDate, util.GetStyleArray())
	if err != nil {
		logger.Error(err.Error())
		return false
	}
	doubaoSeedreamClient := modelapi.NewDoubaoSeedreamClient(flag.DoubaoSeedreamAPIKey)
	firstImageUrl := ""
	for i, chapter := range story.Chapters {
		imgUrl := ""
		err := errors.New("")
		if firstImageUrl != "" {
			chapter.ImagePrompt = "与所给图片中的风格以及人物保持一致，场景无需保持一致。绘制如下场景：" + chapter.ImagePrompt
			imgUrl, err = doubaoSeedreamClient.GenerateImageFromPromptAndGetURL(chapter.ImagePrompt)
		} else {
			imgUrl, err = doubaoSeedreamClient.GenerateImageFromPromptAndImageAndGetURL(chapter.ImagePrompt, firstImageUrl)
		}
		if err != nil {
			logger.Error(err.Error())
			return false
		}
		story.Chapters[i].ImagePath = imgUrl
		voicePath := path.Join(flag.VoiceRoot, uuid.NewString()+".mp3")
		if ok := s.GenerateVoice(chapter.Content, voicePath); ok {
			story.Chapters[i].VoicePath = voicePath
		}

		if firstImageUrl == "" {
			firstImageUrl = imgUrl
		}
	}
	filename := fmt.Sprintf("%s/story_%s.json", flag.StoryRoot, currentDate)
	data, _ := json.MarshalIndent(story, "", "  ")
	result := strings.ReplaceAll(string(data), `\u0026`, "&")
	if _, err := os.Stat(flag.StoryRoot); os.IsNotExist(err) {
		os.MkdirAll(flag.StoryRoot, 0755)
	}
	os.WriteFile(filename, []byte(result), 0644)

	fmt.Printf("成功生成故事: %s\n", story.Title)
	fmt.Printf("背景音乐风格: %s\n", story.MusicStyle)
	fmt.Printf("保存位置: %s\n", filename)

	return true
}

func (s *StoryService) GenerateVoice(text string, filename string) bool {
	if _, err := os.Stat(flag.VoiceRoot); os.IsNotExist(err) {
		os.MkdirAll(flag.VoiceRoot, 0755)
	}
	client := modelapi.NewCosyVoiceClient(flag.CosyVoiceAPIKey, filename)
	err := client.Synthesize([]string{text})
	if err != nil {
		logger.Error(err.Error())
		return false
	}
	return true
}
