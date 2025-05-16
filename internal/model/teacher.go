package model

type QuestionAnalyseRequest struct {
	ImageUrl string `json:"imageUel" binding:"required"` // 图片url
}

type Teacher struct {
	Name string
}
