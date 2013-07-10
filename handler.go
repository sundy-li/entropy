package entropy

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	_ "log"
	"net/http"
	_ "reflect"
)

//处理器接口
type IHandler interface {
	Initialize(rw http.ResponseWriter, req *http.Request)
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

//Base Handler
type Handler struct {
	Response Response
	Request  *http.Request
}

//初始化请求处理器
func (self *Handler) Initialize(rw http.ResponseWriter, req *http.Request) {
	self.Request = req
	self.Response = Response{rw}
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

//请求处理器
type RequestHandler struct {
	Handler
	Application *Application
	flashedMsg  map[string][]string
	tplData     map[string]interface{}
}

//初始化请求处理器
func (self *RequestHandler) InitRequestHandler(app *Application) {
	self.Application = app
	self.flashedMsg = make(map[string][]string)
	self.tplData = make(map[string]interface{})
}

//赋值到模板变量中
func (self *RequestHandler) Assign(name string, value interface{}) {
	self.tplData[name] = value
}

//渲染模板
func (self *RequestHandler) Render(tplPath string) {
	tpl := self.Application.TplEngine.Lookup(tplPath)
	if tpl == nil {
		panic("没有找到指定的模板！")
	}
	d := make(map[string]interface{})
	d["ctx"] = self
	d["vars"] = self.tplData
	self.Response.SetContentType("html")
	tpl.Execute(self.Response, d)
}

//渲染文本
func (self *RequestHandler) RenderText(content string) {
	self.Response.SetContentType("txt")
	fmt.Fprint(self.Response, content)
}

//渲染Json
func (self *RequestHandler) RenderJson(object interface{}) {
	self.Response.SetContentType("json")
	b, _ := json.Marshal(object)
	fmt.Fprint(self.Response, string(b))
}

//设置cookie
func (self *RequestHandler) SetCookie(key, value string, age int) {
	cookie := http.Cookie{Name: key, Value: value, Path: "/", MaxAge: age}
	http.SetCookie(self.Response, &cookie)
}

//获取cookie
func (self *RequestHandler) GetCookie(key string) (string, error) {
	cookie, err := self.Request.Cookie(key)
	if err != nil {
		return "", err
	} else {
		value := cookie.Value
		return value, nil
	}
}

//设置加密cookie,使用aes加密
func (self *RequestHandler) SetSecureCookie(key, value string, age int) {
	AESValue, _ := AesEncrypt([]byte(value), []byte(self.Application.Setting.CookieSecret))
	self.SetCookie(key, base64.StdEncoding.EncodeToString(AESValue), age)
}

//获取加密cookie
func (self *RequestHandler) GetSecureCookie(key string) (string, error) {
	cookie_value, err := self.GetCookie(key)
	if err != nil {
		return "", err
	} else {
		byte_value, _ := base64.StdEncoding.DecodeString(cookie_value)
		value, _ := AesDecrypt(byte_value, []byte(self.Application.Setting.CookieSecret))
		return string(value), nil
	}
}

//刷消息到cookie中
func (self *RequestHandler) flashMessages() {
	byteVlaue, _ := json.Marshal(self.flashedMsg)
	self.SetSecureCookie(self.Application.Setting.FlashCookieName, string(byteVlaue), 0)
}

//刷错误消息
func (self *RequestHandler) FlashError(msg string) {
	self.GetFlashedMessages()
	self.flashedMsg["error"] = append(self.flashedMsg["error"], msg)
	self.flashMessages()
}

//刷成功消息
func (self *RequestHandler) FlashSuccess(msg string) {
	self.GetFlashedMessages()
	self.flashedMsg["success"] = append(self.flashedMsg["success"], msg)
	self.flashMessages()
}

//判断是否有消息被刷
func (self *RequestHandler) HasFlashedMessages() bool {
	self.GetFlashedMessages()
	if len(self.flashedMsg) > 0 {
		return true
	} else {
		return false
	}
}

//更新内部的flashedMsg
func (self *RequestHandler) GetFlashedMessages() error {
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
func (self *RequestHandler) GetFlashedMessagesWithType(msgType string) []string {
	self.GetFlashedMessages()
	rtr := self.flashedMsg[msgType]
	delete(self.flashedMsg, msgType)
	return rtr
}

//请求处理器构造函数
func NewRequestHandler(rw http.ResponseWriter, req *http.Request, app *Application) *RequestHandler {
	return &RequestHandler{Handler: Handler{Response: Response{rw}, Request: req}, Application: app, flashedMsg: make(map[string][]string), tplData: make(map[string]interface{})}
}
