package database

import (
	"fairytale-creator/logger"

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
