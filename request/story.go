package request

type GenerateVoiceReq struct {
	Text     string `json:"text"`
	Filename string `json:"filename"`
}
