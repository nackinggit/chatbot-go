package coze

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"com.imilair/chatbot/bootstrap/config"
	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/pkg/llm/api/base"
	"com.imilair/chatbot/pkg/util"
	"com.imilair/chatbot/pkg/util/ttlmap"
	"com.imilair/chatbot/pkg/xhttp"
	"github.com/golang-jwt/jwt/v5"
)

type CozeObjectString struct {
	Type    string `json:"type,omitempty"`
	Text    string `json:"text,omitempty"`
	FileID  string `json:"file_id,omitempty"`
	FileURL string `json:"file_url,omitempty"`
}

func CozeObjectStringImgObject(fileUrl string) CozeObjectString {
	return CozeObjectString{FileURL: fileUrl, Type: "image"}
}

func CozeObjectStringFileObject(fileUrl string) CozeObjectString {
	return CozeObjectString{FileURL: fileUrl, Type: "file"}
}

func CozeObjectStringTextObject(text string) CozeObjectString {
	return CozeObjectString{Text: text}
}

type CozeMessage struct {
	Role        string `json:"role"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
}

func CozeMessageMultimodalMessage(coss []CozeObjectString) *CozeMessage {
	cossList := make([]map[string]interface{}, len(coss))
	for i, co := range coss {
		cossList[i] = util.Struct2Map(co)
	}
	return &CozeMessage{
		ContentType: "object_string",
		Content:     util.JsonString(cossList),
	}
}

func InitApi(cfg *config.LLMConfig) base.LLMApi {
	return &CozeApi{
		cfg:    cfg,
		client: xhttp.Client{},
	}
}

var cache = ttlmap.New(100, 86400)

func (c *CozeApi) payload() map[string]any {
	return map[string]any{
		"iss": c.cfg.ApiKey,
		"aud": "api.coze.cn",
		"iat": int(time.Now().UnixMilli()),
		"exp": int(time.Now().UnixMilli() + 86400),
		"jti": util.NewSnowflakeID(),
	}
}

func (c *CozeApi) accessToken() (string, error) {
	pk := func() *rsa.PrivateKey {
		b, _ := pem.Decode([]byte(private_key))
		if b == nil {
			return nil
		}
		pk, err := x509.ParsePKCS8PrivateKey(b.Bytes)
		if err != nil {
			xlog.Warnf("parse private key error: %v", err)
			return nil
		}
		return pk.(*rsa.PrivateKey)
	}
	token, err := (&jwt.Token{
		Header: map[string]interface{}{
			"typ": "JWT",
			"alg": jwt.SigningMethodRS256.Alg(),
			"kid": pubkey,
		},
		Claims: jwt.MapClaims(c.payload()),
		Method: jwt.SigningMethodRS256,
	}).SignedString(pk())
	// jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims(c.payload()), func(t *jwt.Token) {
	// 	t.Header["kid"] = pubkey
	// })
	if err != nil {
		xlog.Warnf("sign jwt error: %v", err)
		return "", err
	}
	url := "https://api.coze.cn/api/permission/oauth2/token"
	var resp struct {
		AccessToken string `json:"access_token"`
	}
	headers := map[string][]string{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}
	body := map[string]any{
		"duration_seconds": 86399,
		"grant_type":       "urn:ietf:params:oauth:grant-type:jwt-bearer",
	}
	jsonRequest := xhttp.JsonPostRequest{
		Headers: headers,
		Body:    body,
		Url:     url,
	}
	err = c.client.PostJsonAndBind(&jsonRequest, &resp)
	if err != nil {
		return "", err
	}
	if resp.AccessToken == "" {
		return "", errors.New("access_token is empty")
	}
	return resp.AccessToken, nil
}

func (c *CozeApi) getAccessToken() (string, error) {
	if token, ok := cache.Get("token"); ok {
		return token.(string), nil
	}
	token, err := c.accessToken()
	if token != "" {
		cache.Put("token", token)
	}
	return token, err
}

type ChatResp struct {
	Id             string `json:"id"`
	ConversationId string `json:"conversation_id"`
}

func (c *CozeApi) createChat(botId string, messages []*CozeMessage, variables map[string]any) (*ChatResp, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}
	url := "https://api.coze.cn/v3/chat"
	headers := map[string][]string{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}
	body := map[string]any{
		"bot_id":              botId,
		"user_id":             "imilair",
		"additional_messages": messages,
		"custom_variables":    variables,
	}

	var resp struct {
		Data *ChatResp `json:"data"`
	}
	request := &xhttp.JsonPostRequest{
		Url:     url,
		Headers: headers,
		Body:    body,
	}
	err = c.client.PostJsonAndBind(request, &resp)
	if err != nil {
		xlog.Warnf("Create chat failed: %v", err)
		return nil, err
	}
	return resp.Data, nil
}

func (c *CozeApi) queryChatStatus(chat *ChatResp) (any, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}
	headers := map[string][]string{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}
	url := fmt.Sprintf("https://api.coze.cn/v3/chat/retrieve?chat_id=%s&conversation_id=%s", chat.Id, chat.ConversationId)
	var resp struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	err = c.client.GetAndBind(context.Background(), url, headers, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data.Status, nil
}

func (c *CozeApi) extractText(cresp *ChatResp) (string, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return "", err
	}
	headers := map[string][]string{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}
	url := fmt.Sprintf("https://api.coze.cn/v3/chat/message/list?chat_id=%s&conversation_id=%s", cresp.Id, cresp.ConversationId)
	var resp struct {
		Data []struct {
			Role    string `json:"role"`
			Type    string `json:"type"`
			Content string `json:"content"`
		} `json:"data"`
	}
	err = c.client.GetAndBind(context.Background(), url, headers, &resp)
	if err != nil {
		return "", err
	}
	if len(resp.Data) <= 0 {
		xlog.Warnf("返回结果为空, req: %s, %s", err, cresp.Id, cresp.ConversationId)
		return "", fmt.Errorf("返回结果为空")
	}
	for _, v := range resp.Data {
		if v.Role == "assistant" && v.Type == "answer" {
			return v.Content, nil
		}
	}
	return "我对这个问题不是很了解", nil
}
