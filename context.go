package entropy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

//会话结构体
type Context struct {
	App          *Application
	Req          *http.Request
	Resp         Response
	HandlerName  string
	HandlerCName string
	Flash        *Flash
	Session      *Session
	Data         map[string]interface{}
	startTime    time.Time
	RequireXsrf  bool
	Xsrf         string
	Form         *Form
}

type Flash struct {
	Success string
	Error   string
}

//会话构造函数
func NewContext(app *Application, req *http.Request, rw http.ResponseWriter) *Context {
	return &Context{
		App:         app,
		Req:         req,
		Resp:        Response{rw},
		Flash:       &Flash{},
		Data:        make(map[string]interface{}, 0),
		startTime:   time.Now(),
		RequireXsrf: true,
		Xsrf:        "",
	}
}

func (self *Context) prepareSession() {
	self.Session = &Session{
		SessionId: fmt.Sprintf("%d", time.Now().Nanosecond()),
		store:     NewCookieSession(self.App.Setting.SessionCookieName, self),
	}
	self.Session.Restore()
}

func (self *Context) flushSession() {
	self.Session.Flush()
}

func (self *Context) Assign(name string, value interface{}) {
	self.Data[name] = value
}

func (self *Context) GetStartTime() time.Time {
	return self.startTime
}

func (self *Context) GetXsrf() string {
	return self.Xsrf
}

func (self *Context) HasQueryArgs() bool {
	if len(self.Req.Form) == 0 {
		return true
	} else {
		return false
	}
}

func (self *Context) GetQueryArg(name, defaultValue string) string {
	if param, ok := self.Req.Form[name]; ok {
		return param[0]
	} else {
		return defaultValue
	}
}

//reverse
func (self *Context) Reverse(name string, arg ...interface{}) string {
	if strings.Contains(name, ".") {
		_tmp := strings.Split(name, ".")
		if spec, ok := self.App.Blueprints[_tmp[0]].NamedHandlers[_tmp[1]]; ok {
			url, err := spec.UrlSetParams(arg...)
			if err != nil {
				return err.Error()
			} else {
				return self.App.Blueprints[_tmp[0]].Prefix + url[1:]
			}
		}
	}
	if spec, ok := self.App.NamedHandlers[name]; ok {
		url, err := spec.UrlSetParams(arg...)
		if err != nil {
			return err.Error()
		} else {
			return url[1:]
		}
	}
	return fmt.Sprintf("处理器 %s 没有找到", name)
}

func (self *Context) generateXsrf() {
	if self.RequireXsrf {
		self.Xsrf = base64.StdEncoding.EncodeToString([]byte(time.Now().Format(time.RFC3339)))[22:30] + randString(8)
		self.SetSecureCookie(self.App.Setting.XsrfCookie, self.Xsrf, 600)
	}
}

func (self *Context) IsAjax() bool {
	if value, ajax := self.Req.Header["X-Requested-With"]; ajax {
		if value[0] == "XMLHttpRequest" {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

func (self *Context) FlashError(msg string) {
	self.Flash.Error = msg
}

func (self *Context) FlashSuccess(msg string) {
	self.Flash.Success = msg
}

func (self *Context) IsGet() bool {
	if strings.ToUpper(self.Req.Method) == "GET" {
		return true
	} else {
		return false
	}
}

func (self *Context) IsPost() bool {
	if strings.ToUpper(self.Req.Method) == "POST" {
		return true
	} else {
		return false
	}
}

func (self *Context) restoreMessages() {
	_tmp, err := self.SecureCookie(self.App.Setting.FlashCookieName)
	defer func() {
		self.SetSecureCookie(self.App.Setting.FlashCookieName, "", -1)
	}()
	if err == nil {
		if err := json.Unmarshal([]byte(_tmp), &self.Flash); err != nil {
			log.Println("restoreMessages", err)
		}
	}
}

func (self *Context) flushMessage() {
	_tmp, err := json.Marshal(self.Flash)
	if err == nil {
		self.SetSecureCookie(self.App.Setting.FlashCookieName, string(_tmp), 2)
	}
}

//设置加密cookie,使用aes加密
func (self *Context) SetSecureCookie(key, value string, age int) {
	self.SetCookie(key, string(Base64Encode([]byte(value))), age)
}

//获取加密cookie
func (self *Context) SecureCookie(key string) (string, error) {
	cookie_value, err := self.Cookie(key)
	if err != nil {
		log.Println("get cookie", err)
		return "", err
	} else {
		log.Println("before base64", cookie_value)
		value, err := Base64Decode([]byte(cookie_value))
		log.Println("after base64", value)
		return string(value), err
	}
}

//设置cookie
func (self *Context) SetCookie(key, value string, age int) {
	cookie := http.Cookie{Name: key, Value: value, Path: "/"}
	if age != 0 {
		cookie.MaxAge = age
	}
	http.SetCookie(self.Resp, &cookie)
}

//获取cookie
func (self *Context) Cookie(key string) (string, error) {
	cookie, err := self.Req.Cookie(key)
	if err != nil {
		return "", err
	} else {
		value := cookie.Value
		return value, nil
	}
}

func (self *Context) Html(tplName string) Result {
	return NewHtmlResult(self, tplName)
}
