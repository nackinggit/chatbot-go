package model

type ExtractNameRequest struct {
	Content []string `json:"content" binding:"required" err:"content不能为空"`
}

type ExtractNameResponse struct {
	Nickname string `json:"nickname"`
}

type CommentPicRequest struct {
	PicUrl string `json:"picUrl" binding:"required" err:"picUrl不能为空"`
}

type CommentPicResponse struct {
	Comment string `json:"content"`
}
