package service

import (
	"fairytale-creator/database"
	"fairytale-creator/flag"
	"fairytale-creator/logger"
	"fairytale-creator/modelapi"
	"fairytale-creator/response"
	"fairytale-creator/util"
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

func (s *StoryService) GenerateStory() *response.Story {
	currentDate := time.Now().Format("2006-01-02")
	theme := util.GenerateDailyTheme(currentDate)
	client := modelapi.NewDeepSeekClient(s.DeepSeekAPIKey, s.DeepSeekUrl)
	story, err := client.GenerateFairyTale(theme, currentDate, util.GetStyleArray())
	if err != nil {
		logger.Error(err.Error())
		return nil
	}
	logger.Log("deepseek end", story.Description)
	// return story
	doubaoSeedreamClient := modelapi.NewDoubaoSeedreamClient(flag.DoubaoSeedreamAPIKey)
	firstImageUrl := ""
	for i, chapter := range story.Chapters {
		imgUrl := ""
		if firstImageUrl != "" {
			imgUrl, err = doubaoSeedreamClient.GenerateImageFromPromptAndGetURL(chapter.ImagePrompt)
		} else {
			imgUrl, err = doubaoSeedreamClient.GenerateImageFromPromptAndImageAndGetURL(chapter.ImagePrompt, firstImageUrl)
		}
		if err != nil {
			logger.Error(err.Error())
			return nil
		}
		story.Chapters[i].ImagePath = imgUrl
		logger.Log("chapter image", imgUrl)
		voicePath := path.Join(flag.VoiceRoot, uuid.NewString()+currentDate+".mp3")
		if ok := s.GenerateVoice(chapter.Content, voicePath); ok {
			story.Chapters[i].VoicePath = voicePath
		}

		if firstImageUrl == "" {
			firstImageUrl = imgUrl
		}
		logger.Log("chapter voice", voicePath)
	}
	// filename := fmt.Sprintf("%s/story_%s.json", flag.StoryRoot, currentDate)
	// data, _ := json.MarshalIndent(story, "", "  ")
	// result := strings.ReplaceAll(string(data), `\u0026`, "&")
	// if _, err := os.Stat(flag.StoryRoot); os.IsNotExist(err) {
	// 	os.MkdirAll(flag.StoryRoot, 0755)
	// }
	// os.WriteFile(filename, []byte(result), 0644)

	// fmt.Printf("成功生成故事: %s\n", story.Title)
	// fmt.Printf("背景音乐风格: %s\n", story.MusicStyle)
	// fmt.Printf("保存位置: %s\n", filename)

	return story
}

func (s *StoryService) AddStory(story *response.Story) error {
	currentDate := time.Now().Format("2006-01-02")
	storyDao := database.NewStoryDao()
	storyModel := database.Story{
		Title:       story.Title,
		Author:      story.Author,
		Description: story.Description,
		MusicStyle:  story.MusicStyle,
		Status:      0,
	}
	response, err := storyDao.AddStoryToD1(&storyModel)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	storyModel.ID = uint(response.Result[0].Meta.LastRowID)
	chapterDao := database.NewChapterDao()
	for _, chapter := range story.Chapters {
		imageName := uuid.NewString() + currentDate + ".png"
		uploader, err := modelapi.NewR2Uploader(flag.CfAccountID, flag.R2AccessKeyID, flag.R2AccessKeySecret, "fairytale")
		if err != nil {
			logger.Error(err.Error())
			return err
		}
		err = uploader.UploadFromURL(chapter.ImagePath, imageName)
		if err != nil {
			logger.Error(err.Error())
			return err
		}
		voiceTemp := strings.Split(chapter.VoicePath, "/")
		voiceName := voiceTemp[len(voiceTemp)-1]
		err = uploader.UploadFromLocalFile(chapter.VoicePath, voiceName)
		if err != nil {
			logger.Error(err.Error())
			return err
		}
		chapterModel := database.Chapter{
			StoryID:     storyModel.ID,
			Title:       chapter.Title,
			Content:     chapter.Content,
			ImagePrompt: chapter.ImagePrompt,
			ImagePath:   imageName,
			VoicePath:   voiceName,
		}
		_, err = chapterDao.AddChapterToD1(&chapterModel)
		if err != nil {
			logger.Error(err.Error())
			return err
		}
	}
	return nil
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
