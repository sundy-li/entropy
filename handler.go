package entropy

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

//处理器接口
type IHandler interface {
	Initialize(rw http.ResponseWriter, req *http.Request, app *Application)
	Prepare()
	Get()
	Post()
	Finish()
	Head()
	Delete()
	Patch()
	Put()
	Options()
}

//请求处理器
type Handler struct {
	Response     http.ResponseWriter
	Request      *http.Request
	Session      *Session
	Application  *Application
	flashedMsg   map[string][]string
	tplData      map[string]interface{}
	BeforeRender func()
}

//初始化请求处理器
func (self *Handler) Initialize(rw http.ResponseWriter, req *http.Request, app *Application) {
	self.Request = req
	self.Response = rw
	self.Application = app
	self.flashedMsg = make(map[string][]string)
	self.tplData = make(map[string]interface{})
	self.Session = &Session{
		store: NewCookieSession(app.Setting.SessionCookieName, self),
	}
	//这里是一个坑，待填
	// self.BeforeRender = func() {
	// 	self.Session.Flush()
	// }
}

func (self *Handler) Prepare() {

}

func (self *Handler) Get() {
	panic(errors.New("GET method is not implemented"))
}

func (self *Handler) Post() {
	panic(errors.New("POST method is not implemented"))
}

func (self *Handler) Finish() {
	self.Request.Close = true
}

func (self *Handler) Head() {
	panic(errors.New("HEAD method is not implemented"))
}

func (self *Handler) Delete() {
	panic(errors.New("DELETE method is not implemented"))
}

func (self *Handler) Patch() {
	panic(errors.New("PATHC method is not implemented"))
}

func (self *Handler) Put() {
	panic(errors.New("PUT method is not implemented"))
}

func (self *Handler) Options() {
	panic(errors.New("OPTIONS method is not implemented"))
}

//跳转
func (self *Handler) Redirect(url string, permanent bool) {
	var status int
	if permanent {
		status = 301
	} else {
		status = 302
	}
	self.Response.Header().Set("Location", url)
	self.Response.WriteHeader(status)
}

//reverse
func (self *Handler) Reverse(name string, arg ...interface{}) string {
	if spec, ok := self.Application.NamedHandlers[name]; ok {
		url, err := spec.UrlSetParams(arg...)
		if err != nil {
			return err.Error()
		} else {
			return url
		}
	}
	return fmt.Sprintf("处理器 %s 没有找到", name)
}

//赋值到模板变量中
func (self *Handler) Assign(name string, value interface{}) {
	self.tplData[name] = value
}

//渲染模板
func (self *Handler) Render(tplPath string) {
	tpl := self.Application.TplEngine.Lookup(tplPath)
	if tpl == nil {
		panic("没有找到指定的模板！")
	}
	d := make(map[string]interface{})
	d["xsrf"] = fmt.Sprintf("%x", sha1.New().Sum([]byte(time.Now().Format(time.RFC3339))))
	d["ctx"] = self
	d["vars"] = self.tplData
	self.Response.Header().Set("Content-Type", "text/html")
	tpl.Execute(self.Response, d)
}

//渲染文本
func (self *Handler) RenderText(content string) {
	self.Response.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(self.Response, content)
}

//渲染Json
func (self *Handler) RenderJson(object interface{}) {
	self.Response.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(object)
	fmt.Fprint(self.Response, string(b))
}

//设置cookie
func (self *Handler) SetCookie(key, value string, age int) {
	cookie := http.Cookie{Name: key, Value: value, Path: "/"}
	if age != 0 {
		cookie.MaxAge = age
	}
	http.SetCookie(self.Response, &cookie)
}

//获取cookie
func (self *Handler) GetCookie(key string) (string, error) {
	cookie, err := self.Request.Cookie(key)
	if err != nil {
		return "", err
	} else {
		value := cookie.Value
		return value, nil
	}
}

//设置加密cookie,使用aes加密
func (self *Handler) SetSecureCookie(key, value string, age int) {
	AESValue, e := AesEncrypt([]byte(value), []byte(self.Application.Setting.Secret))
	if e != nil {
		panic(e.Error())
	}
	self.SetCookie(key, base64.StdEncoding.EncodeToString(AESValue), age)
}

//获取加密cookie
func (self *Handler) GetSecureCookie(key string) (string, error) {
	cookie_value, err := self.GetCookie(key)
	if err != nil {
		return "", err
	} else {
		byte_value, _ := base64.StdEncoding.DecodeString(cookie_value)
		value, _ := AesDecrypt(byte_value, []byte(self.Application.Setting.Secret))
		return string(value), nil
	}
}

//刷消息到cookie中
func (self *Handler) flashMessages() {
	byteVlaue, _ := json.Marshal(self.flashedMsg)
	self.SetSecureCookie(self.Application.Setting.FlashCookieName, string(byteVlaue), 0)
}

//刷错误消息
func (self *Handler) FlashError(msg string) {
	//self.GetFlashedMessages()
	self.flashedMsg["error"] = append(self.flashedMsg["error"], msg)
	self.flashMessages()
}

//刷成功消息
func (self *Handler) FlashSuccess(msg string) {
	//self.GetFlashedMessages()
	self.flashedMsg["success"] = append(self.flashedMsg["success"], msg)
	self.flashMessages()
}

//判断是否有消息被刷
func (self *Handler) HasFlashedMessages(msgType string) bool {
	self.GetFlashedMessages()
	if len(self.flashedMsg[msgType]) > 0 {
		return true
	} else {
		return false
	}
}

//更新内部的flashedMsg
func (self *Handler) GetFlashedMessages() error {
	value, err := self.GetSecureCookie(self.Application.Setting.FlashCookieName)
	if err != nil {
		return err
	} else {
		err := json.Unmarshal([]byte(value), &self.flashedMsg)
		if err != nil {
			return err
		} else {
			return nil
		}
	}
}

//获取消息,供模板使用
func (self *Handler) GetFlashedMessagesWithType(msgType string) []string {
	self.GetFlashedMessages()
	rtr := self.flashedMsg[msgType]
	delete(self.flashedMsg, msgType)
	self.flashMessages()
	return rtr
}
