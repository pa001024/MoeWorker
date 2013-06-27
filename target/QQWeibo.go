package target

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	TQQ_OAUTH_VERSION = "2.a"
)

type QQWeibo struct { // 腾讯微博API 实现接口IWeibo
	IWeibo

	AppKey       string    `json:client_id`     // AppKey
	AppSecret    string    `json:client_secret` // AppSecret
	CallbackUrl  string    `json:redirect_uri`  // 验证URL
	Token        string    `json:access_token`  // OAuth2.0 验证码
	ExpiresIn    time.Time `json:expires_in`    // 失效时间
	RefreshToken string    `json:refresh_token` // OAuth2.0 疼迅的二次验证码
	OpenID       string    `json:openid`        // 平台ID
	// ClientIP     string    `json:clientip`   // 改为动态计算
}
type QQWeiboResult struct {
	ErrorCode int            `json:"errcode"` // 错误代码
	Error     string         `json:"msg"`     // 返回信息
	Ret       int            `json:"ret"`     // 返回值
	Data      *QQWeiboStatus `json:"data"`    // 数据
	// SeqId     string         `json:"seqid"`   // 序列号 (无需使用)
}
type QQWeiboStatus struct {
	IStatus
	Id        int64  `json:"id"`        // 微博id
	CreatedAt string `json:"timestamp"` // 微博发表时间
}

func (this *QQWeibo) Authorize() (authurl string) {
	return "https://open.t.qq.com/cgi-bin/oauth2/authorize?" + (url.Values{
		"client_id":     {this.AppKey},
		"redirect_uri":  {this.CallbackUrl},
		"response_type": {"code"},
	}).Encode()
}
func (this *QQWeibo) AccessToken(code string) (token string) {
	res, err := http.PostForm("https://open.t.qq.com/cgi-bin/oauth2/access_token",
		url.Values{
			"grant_type":    {"authorization_code"},
			"client_id":     {this.AppKey},      // yourappkey
			"client_secret": {this.AppSecret},   // yourpppsecret
			"code":          {code},             // xxxxxxxxxxxxxx
			"redirect_uri":  {this.CallbackUrl}, // http://some/weibocb.php
		})
	if err != nil {
		Log("Fail to AccessToken:", err)
		return
	}

	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		Log("Fail to AccessToken:", err)
		return
	}
	body, _ := url.ParseQuery(string(b))
	if body.Get("error") != "" || body.Get("access_token") == "" {
		Log("Fail to AccessToken(Remote):", body.Get("error"))
		return
	}
	this.Token = body.Get("access_token")
	this.OpenID = body.Get("openid")
	this.RefreshToken = body.Get("refresh_token")
	i, _ := strconv.Atoi(body.Get("expires_in"))
	ex := time.Now().Add(time.Duration(i) * time.Second)
	this.ExpiresIn = ex
	return this.Token
}
func (this *QQWeibo) PostStatus(api string, args *url.Values) (rst *QQWeiboStatus) {
	// OAuth
	args.Set("oauth_consumer_key", this.AppKey)
	args.Set("access_token", this.Token)
	args.Set("openid", this.OpenID)
	args.Set("clientip", IP)
	res, err := http.PostForm("https://open.t.qq.com/api/t/"+api, *args)
	if err != nil {
		Log("Error on call", api+":", err)
		return
	}
	defer res.Body.Close()
	dst := &QQWeiboResult{}
	json.NewDecoder(res.Body).Decode(dst)
	if dst.ErrorCode != 0 {
		Log("Error on call", api+"(Remote):", dst.ErrorCode, ":", dst.Error)
		return nil
	}
	rst = dst.Data
	return
}
func (this *QQWeibo) Update(status string) (rst *QQWeiboStatus) {
	rst = this.PostStatus("add", &url.Values{
		"format":  {"json"},
		"content": {status},
	})
	return
}
func (this *QQWeibo) Repost(status string, oriId int64) (rst *QQWeiboStatus) {
	rst = this.PostStatus("re_add", &url.Values{
		"format":  {"json"},
		"content": {status},
		"reid":    {ToString(oriId)},
	})
	return
}
func (this *QQWeibo) UploadUrl(status string, urlText string) (rst *QQWeiboStatus) {
	rst = this.PostStatus("add_pic_url", &url.Values{
		"format":  {"json"},
		"content": {status},
		"pic_url": {urlText},
	})
	return
}
func (this *QQWeibo) Upload(status string, pic io.Reader) (rst *QQWeiboStatus) {
	// multipart/form-data
	var bpic bytes.Buffer
	formdata := multipart.NewWriter(&bpic)
	formdata.WriteField("oauth_consumer_key", this.AppKey)
	formdata.WriteField("access_token", this.Token)
	formdata.WriteField("openid", this.OpenID)
	formdata.WriteField("clientip", IP)

	formdata.WriteField("format", "json")
	formdata.WriteField("content", status)
	picdata, _ := formdata.CreateFormFile("pic", "image.png")
	io.Copy(picdata, pic)
	formdata.Close()

	res, err := http.Post("https://open.t.qq.com/api/t/add_pic", formdata.FormDataContentType(), &bpic)
	if err != nil {
		Log("Error on call upload :", err)
		return
	}
	defer res.Body.Close()
	dst := &QQWeiboResult{}
	json.NewDecoder(res.Body).Decode(dst)
	if dst.ErrorCode != 0 {
		Log("Error on call upload (Remote):", dst.ErrorCode, ":", dst.Error)
		return nil
	}
	rst = dst.Data
	return
}

func (this *QQWeiboStatus) Url() (urlText string) {
	urlText = fmt.Sprintf("http://t.qq.com/p/t/%v", this.Id)
	return
}