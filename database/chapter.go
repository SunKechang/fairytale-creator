package database

import (
	"fairytale-creator/flag"
	"fairytale-creator/logger"
	"fairytale-creator/modelapi"
	"time"

	"gorm.io/gorm"
)

const ChapterTableName = "chapter"

type Chapter struct {
	gorm.Model
	StoryID     uint   `json:"story_id" gorm:"not null;column:story_id;index"`
	Title       string `json:"title" gorm:"not null;column:title"`
	Content     string `json:"content" gorm:"not null;column:content"`
	ImagePrompt string `json:"image_prompt" gorm:"not null;column:image_prompt"`
	ImagePath   string `json:"image_path" gorm:"not null;column:image_path"`
	VoicePath   string `json:"voice_path" gorm:"not null;column:voice_path"`
}

func (c Chapter) TableName() string {
	return ChapterTableName
}

type ChapterDao struct {
	BaseDao
}

func NewChapterDao() *ChapterDao {
	return &ChapterDao{
		BaseDao{Engine: GetDB()},
	}
}

func (p *ChapterDao) AddChapter(c *Chapter) error {
	q := p.GetDB()
	q.Create(c)
	if q.Error != nil {
		logger.Error("创建章节报错：", q.Error.Error())
		return InterError
	}
	return nil
}

func (p *ChapterDao) AddChapterToD1(c *Chapter) (*modelapi.D1QueryResponse, error) {
	client := modelapi.NewD1Client(flag.CfAccountID, flag.D1DatabaseID, flag.D1APIKey)
	response, err := client.ExecuteQuery("INSERT INTO chapter (story_id, title, content, image_prompt, image_path, voice_path, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		[]interface{}{c.StoryID, c.Title, c.Content, c.ImagePrompt, c.ImagePath, c.VoicePath, time.Now().Unix(), time.Now().Unix(), nil})
	if err != nil {
		logger.Error("添加章节到D1报错：", err.Error())
		return nil, err
	}
	return response, nil
}
