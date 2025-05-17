package model

type QuestionAnalyseRequest struct {
	ImageUrl string `json:"imageUrl" binding:"required" err:"imageUrl is required"` // 图片url
}

type Teacher struct {
	Name string
}
