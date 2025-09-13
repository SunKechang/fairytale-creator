package database

import (
	"fairytale-creator/flag"
	"fairytale-creator/logger"
	"fairytale-creator/modelapi"
	"time"

	"gorm.io/gorm"
)

const StoryTableName = "story"

type Story struct {
	gorm.Model
	Title       string `json:"title" gorm:"not null;column:title"`
	Author      string `json:"author" gorm:"not null;column:author"`
	Description string `json:"description" gorm:"not null;column:description"`
	MusicStyle  string `json:"music_style" gorm:"not null;column:music_style"`
	Status      int    `json:"status" gorm:"not null;column:status"` // 0: 待审阅, 1: 已上传, 2: 生成完成
}

func (s Story) TableName() string {
	return StoryTableName
}

type StoryDao struct {
	BaseDao
}

func NewStoryDao() *StoryDao {
	return &StoryDao{
		BaseDao{Engine: GetDB()},
	}
}

func (p *StoryDao) AddStory(s *Story) error {
	q := p.GetDB()
	q.Create(s)
	if q.Error != nil {
		logger.Error("创建故事报错：", q.Error.Error())
		return InterError
	}
	return nil
}

func (p *StoryDao) AddStoryToD1(s *Story) (*modelapi.D1QueryResponse, error) {
	client := modelapi.NewD1Client(flag.CfAccountID, flag.D1DatabaseID, flag.D1APIKey)
	response, err := client.ExecuteQuery("INSERT INTO story (title, author, description, music_style, status, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		[]interface{}{s.Title, s.Author, s.Description, s.MusicStyle, s.Status, time.Now().Unix(), time.Now().Unix(), nil})
	if err != nil {
		logger.Error("添加故事到D1报错：", err.Error())
		return nil, err
	}
	return response, nil
}
