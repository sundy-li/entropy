package entropy

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
)

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
	RespWriter http.ResponseWriter
	Request    *http.Request
}

func (self *Handler) Initialize(rw http.ResponseWriter, req *http.Request) {
	self.Request = req
	self.RespWriter = rw
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

func (self *Handler) Redirect(url string, permanent bool) {
	var status int
	if permanent {
		status = 301
	} else {
		status = 302
	}
	self.RespWriter.Header().Set("Location", url)
	self.RespWriter.WriteHeader(status)
}

type RequestHandler struct {
	Handler
	Application *Application
	Transformer http.ResponseWriter
	flashedMsg  map[string][]string
	tplData     map[string]interface{}
}

func (self *RequestHandler) InitRequestHandler(app *Application, transformer http.ResponseWriter) {
	self.Application = app
	self.Transformer = transformer
	self.flashedMsg = make(map[string][]string)
	self.tplData = make(map[string]interface{})
}

func (self *RequestHandler) Assign(name string, value interface{}) {
	self.tplData[name] = value
}

func (self *RequestHandler) SetCookie(key, value string, age int) {
	cookie := http.Cookie{Name: key, Value: value, Path: "/", MaxAge: age}
	http.SetCookie(self.RespWriter, &cookie)
}

func (self *RequestHandler) GetCookie(key string) (string, error) {
	cookie, err := self.Request.Cookie(key)
	if err != nil {
		return "", err
	} else {
		value := cookie.Value
		return value, nil
	}
}

func (self *RequestHandler) SetSecureCookie(key, value string, age int) {
	AESValue, _ := AesEncrypt([]byte(value), []byte(self.Application.Setting.CookieSecret))
	self.SetCookie(key, base64.StdEncoding.EncodeToString(AESValue), age)
}

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

func (self *RequestHandler) flashMessages() {
	byteVlaue, _ := json.Marshal(self.flashedMsg)
	self.SetSecureCookie(self.Application.Setting.FlashCookieName, string(byteVlaue), 0)
}

func (self *RequestHandler) FlashError(msg string) {
	self.GetFlashedMessages()
	self.flashedMsg["error"] = append(self.flashedMsg["error"], msg)
	self.flashMessages()
}

func (self *RequestHandler) FlashSuccess(msg string) {
	self.GetFlashedMessages()
	self.flashedMsg["success"] = append(self.flashedMsg["success"], msg)
	self.flashMessages()
}

func (self *RequestHandler) HasFlashedMessages() bool {
	self.GetFlashedMessages()
	if len(self.flashedMsg) > 0 {
		return true
	} else {
		return false
	}
}

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

func (self *RequestHandler) GetFlashedMessagesWithType(msgType string) []string {
	self.GetFlashedMessages()
	rtr := self.flashedMsg[msgType]
	delete(self.flashedMsg, msgType)
	return rtr
}

//请求处理器构造函数
func NewRequestHandler(rw http.ResponseWriter, req *http.Request, app *Application) *RequestHandler {
	return &RequestHandler{Handler: Handler{rw, req}, Application: app, flashedMsg: make(map[string][]string), tplData: make(map[string]interface{})}
}

//跳转处理器
type RedirectHandler struct {
	Handler
	destinationUrl string
	isForever      bool
}

func (self *RedirectHandler) Get() {
	self.Redirect(self.destinationUrl, self.isForever)
}

func NewRedirectHandler(rw http.ResponseWriter, req *http.Request, destinationUrl string, permanet bool) *RedirectHandler {
	return &RedirectHandler{Handler: Handler{rw, req}, destinationUrl: destinationUrl, isForever: permanet}
}
