package imapi

import (
	"context"
	"errors"
	"fmt"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/service"
	"com.imilair/chatbot/internal/service/config"
	"com.imilair/chatbot/pkg/util"
	"com.imilair/chatbot/pkg/xhttp"
)

var ImapiService *imapi

type imapi struct {
	cfg     *config.ImApi
	baseUrl string
}

func (t *imapi) Name() string {
	return "serivice.imapi"
}

func (t *imapi) InitAndStart() (err error) {
	xlog.Infof("init service `%s`", t.Name())
	cfg := service.Config.ImApi
	if err := cfg.Validate(); err != nil {
		xlog.Warnf("`%s` config error: %v", t.Name(), err)
		return err
	}
	t.cfg = cfg
	t.baseUrl = cfg.BaseUrl
	xlog.Infof("`%s` inited", t.Name())
	ImapiService = t
	return nil
}

func (t *imapi) Stop() {
	xlog.Infof("stop service `%s`", t.Name())
}

func init() {
	service.Register(&imapi{})
}

func (t *imapi) SendMessage(reply *ReplyMessage, scene string) error {
	if scene == "" {
		scene = "chat"
	}

	target_id := reply.ReplyTo.TargetId
	url := fmt.Sprintf("%s/chat/add", t.baseUrl)
	headers := map[string][]string{
		"LoginUserId": {reply.SenderId},
	}
	body := map[string]any{
		"scene_id": target_id,
		"scene":    scene,
		"type":     reply.ReplyContent.Type,
		"content": util.JsonString(map[string]any{
			string(reply.ReplyContent.Type): reply.ReplyContent.Content,
			"srcContentId":                  reply.ReplyContent.SrcContentId,
		}),
	}
	req := &xhttp.JsonPostRequest{
		Url:     url,
		Headers: headers,
		Body:    body,
	}
	var resp map[string]any = map[string]any{}
	err := xhttp.PostJsonAndBind(req, &resp)
	if err != nil {
		xlog.Warnf("发送消息失败: %v", err)
	}
	return err
}

// 发布评论
func (t *imapi) Comment(senderId string, postId int, content string, commentId int, replyUserId string) {
	url := fmt.Sprintf("%s/user/comment", t.baseUrl)
	headers := map[string][]string{
		"LoginUserId": {senderId},
	}
	body := map[string]any{
		"post_id":        postId,
		"content":        content,
		"comment_id":     commentId,
		"replay_user_id": replyUserId,
	}
	req := &xhttp.JsonPostRequest{
		Url:     url,
		Headers: headers,
		Body:    body,
	}
	var resp map[string]any = map[string]any{}
	err := xhttp.PostJsonAndBind(req, &resp)
	if err != nil {
		xlog.Warnf("发送消息失败: %v", err)
	}
}

func (t *imapi) QueryChatContent(imBotId string, msgId string) (*ChatContent, error) {
	url := fmt.Sprintf("%s/chat/info?chat_id=%s", t.baseUrl, msgId)
	headers := map[string][]string{
		"LoginUserId": {imBotId},
	}
	var resp map[string]any = map[string]any{}
	err := xhttp.GetAndBind(context.Background(), url, headers, &resp)
	if err != nil {
		xlog.Warnf("查询消息内容失败: imbotId: %s, msgId: %d, err: %v", imBotId, msgId, err)
		return nil, err
	}
	xlog.Debugf("查询到的消息内容: imbotId: %s, msgId: %d, resp: %v", imBotId, msgId, util.JsonString(resp))
	if data, ok := resp["data"].(map[string]any); ok {
		msgType := data["type"].(string)
		cc := ChatContent{Type: msgType}
		if content, ok := data["content"].(string); ok {
			err := util.Unmarshal([]byte(content), &cc)
			if err != nil {
				xlog.Warnf("解析消息内容失败: imbotId: %s, msgId: %d, err: %v", imBotId, msgId, err)
				return nil, err
			}
			return &cc, nil
		}
	}
	xlog.Warnf("无法解析消息内容: imbotId: %s, msgId: %d", imBotId, msgId)
	return nil, errors.New("无法解析消息内容")
}

func (t *imapi) QueryPostComments(postId string) (*PostComments, error) {
	url := fmt.Sprintf("%s/comment/list?post_id=%s", t.baseUrl, postId)
	headers := map[string][]string{"LoginUserId": {"7"}}
	var resp struct {
		Code int           `json:"code"`
		Msg  string        `json:"msg"`
		Data *PostComments `json:"data"`
	}
	err := xhttp.GetAndBind(context.Background(), url, headers, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Data == nil {
		xlog.Warnf("获取评论信息失败: %v", util.JsonString(resp))
	}
	return resp.Data, nil
}

func (t *imapi) QueryPostById(id string) (*ImPost, error) {
	url := fmt.Sprintf("%s/post/info?post_id=%s", t.baseUrl, id)
	headers := map[string][]string{"LoginUserId": {"7"}}
	var resp struct {
		Code int     `json:"code"`
		Msg  string  `json:"msg"`
		Data *ImPost `json:"data"`
	}
	err := xhttp.GetAndBind(context.Background(), url, headers, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Data == nil {
		xlog.Warnf("获取评论信息失败: %v", util.JsonString(resp))
	}
	return resp.Data, nil
}

// 获取评论信息
func (t *imapi) QueryCommentById(commentId string) (*ImComment, error) {
	url := fmt.Sprintf("%s/comment/info?comment_id=%s", t.baseUrl, commentId)
	headers := map[string][]string{
		"LoginUserId": {"7"},
	}
	var resp struct {
		Code int        `json:"code"`
		Msg  string     `json:"msg"`
		Data *ImComment `json:"data"`
	}
	err := xhttp.GetAndBind(context.Background(), url, headers, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Data == nil {
		xlog.Warnf("获取评论信息失败: %v", util.JsonString(resp))
	}
	return resp.Data, nil
}

func (t *imapi) QueryChatRoomSetting(ctx context.Context, roomId string) (*ChatRoomSetting, error) {
	url := fmt.Sprintf("%s/room/info?room_id=%s", t.baseUrl, roomId)
	headers := map[string][]string{
		"LoginUserId": {"7"},
	}
	var resp struct {
		Code int              `json:"code"`
		Msg  string           `json:"msg"`
		Data *ChatRoomSetting `json:"data"`
	}
	err := xhttp.GetAndBind(ctx, url, headers, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Data == nil {
		xlog.Warnf("获取评论信息失败: %v", util.JsonString(resp))
	}
	return resp.Data, nil
}
