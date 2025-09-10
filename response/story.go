package response

type Story struct {
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	Description string    `json:"description"`
	Chapters    []Chapter `json:"chapters"`
	MusicStyle  string    `json:"music_style"` // 背景音乐风格描述
	CreatedAt   string    `json:"created_at"`  // 创建日期，用于确保唯一性
}

type Chapter struct {
	Title         string `json:"title"`
	Content       string `json:"content"`
	ImagePrompt   string `json:"image_prompt"`
	ChapterNumber int    `json:"chapter_number"`
	ImagePath     string `json:"image_path,omitempty"`
	VoicePath     string `json:"voice_path,omitempty"`
}
