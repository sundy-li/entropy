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

type Context struct {
	App       *Application
	Req       *http.Request
	Resp      Response
	Messages  map[string][]string
	Data      map[string]interface{}
	startTime time.Time
	Xsrf      string
	Form      *Form
}

func NewContext(app *Application, req *http.Request, rw http.ResponseWriter) *Context {
	return &Context{
		App:       app,
		Req:       req,
		Resp:      Response{rw},
		Messages:  make(map[string][]string, 0),
		Data:      make(map[string]interface{}, 0),
		startTime: time.Now(),
		Xsrf:      "",
	}
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

func (self *Context) GetArg(name, defaultValue string) string {
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
	self.Xsrf = base64.StdEncoding.EncodeToString([]byte(time.Now().Format(time.RFC3339)))[22:30] + randString(8)
	self.SetSecureCookie(XSRF, self.Xsrf, 3600)
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

func (self *Context) Flash(key, msg string) {
	self.Messages[key] = append(self.Messages[key], msg)
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
	_tmp, err := self.GetSecureCookie(self.App.Setting.FlashCookieName)
	if err == nil {
		if err := json.Unmarshal([]byte(_tmp), &self.Messages); err != nil {
			log.Println(err)
		}
	}
}

func (self *Context) flushMessage() {
	_tmp, err := json.Marshal(self.Messages)
	if err == nil {
		self.SetSecureCookie(self.App.Setting.FlashCookieName, string(_tmp), 2)
	}
}

//设置加密cookie,使用aes加密
func (self *Context) SetSecureCookie(key, value string, age int) {
	AESValue, e := AesEncrypt([]byte(value), []byte(self.App.Setting.Secret))
	if e != nil {
		panic(e.Error())
	}
	self.SetCookie(key, base64.StdEncoding.EncodeToString(AESValue), age)
}

//获取加密cookie
func (self *Context) GetSecureCookie(key string) (string, error) {
	cookie_value, err := self.GetCookie(key)
	if err != nil {
		return "", err
	} else {
		byte_value, err := base64.StdEncoding.DecodeString(cookie_value)
		if err != nil {
			log.Println("getsecurecookie", err.Error())
		}
		value, err := AesDecrypt(byte_value, []byte(self.App.Setting.Secret))
		if err != nil {
			log.Println("getsecurecookie", err.Error())
		}
		return string(value), nil
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
func (self *Context) GetCookie(key string) (string, error) {
	cookie, err := self.Req.Cookie(key)
	if err != nil {
		return "", err
	} else {
		value := cookie.Value
		return value, nil
	}
}
