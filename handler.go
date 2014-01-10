package entropy

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
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
	GetStartTime() time.Time
}

//请求处理器
type Handler struct {
	startTime   time.Time
	Response    Response
	Request     *http.Request
	Session     *Session
	Application *Application
	errorMsg    map[string]string
	successMsg  map[string]string
	tplData     map[string]interface{}
	Form        *EForm
}

//初始化请求处理器
func (self *Handler) Initialize(rw http.ResponseWriter, req *http.Request, app *Application) {
	self.startTime = time.Now()
	self.Request = req
	self.Response = Response{rw}
	self.Application = app
	self.errorMsg = make(map[string]string)
	self.successMsg = make(map[string]string)
	self.tplData = make(map[string]interface{})
	self.Session = &Session{
		store: NewCookieSession(app.Setting.SessionCookieName, self),
	}
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

func (self *Handler) RestoreSession() {
	self.Session.Restore()
}

func (self *Handler) FlushSession() {
	self.Session.Flush()
}

func (self *Handler) GetStartTime() time.Time {
	return self.startTime
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

//xsrf 验证
func (self *Handler) ValidXsrf() bool {
	xsrfRaw, okRaw := self.Session.GetSession(self.Application.Setting.XsrfKey).(string)
	//当从session中读取到xsrf值并且该值与用户提交表单中的xsrf一致时,返回true
	//其他情况均返回false
	if okRaw && (self.Form.Xsrf == xsrfRaw) {
		return true
	} else {
		return false
	}
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

func (self *Handler) RenderImage(img image.Image, imgType int) {
	self.FlushSession()
	b := bufio.NewWriter(self.Response.ResponseWriter)
	switch imgType {
	case IMAGEPNG:
		{
			png.Encode(b, img)
			self.Response.SetContentType("png")
		}
	case IMAGEGIF:
		{
			gif.Encode(b, img, nil)
			self.Response.SetContentType("gif")
		}
	case IMAGEJPEG:
		{
			jpeg.Encode(b, img, nil)
			self.Response.SetContentType("jpeg")
		}
	default:
		panic("错误的图片类型!")
	}
	b.Flush()
}

//渲染模板
func (self *Handler) Render(tplPath string) {
	tpl := self.Application.TplEngine.Lookup(tplPath)
	tpl.Funcs(self.Application.TplFuncs)
	if tpl == nil {
		panic("没有找到指定的模板！")
	}
	d := make(map[string]interface{})
	if self.Form != nil {
		self.Form.Xsrf = base64.StdEncoding.EncodeToString([]byte(time.Now().Format(time.RFC3339)))[22:30] + randString(8)
		self.Session.SetSession(self.Application.Setting.XsrfKey, self.Form.Xsrf)
		d["form"] = self.Form
	}
	self.FlushSession()
	if self.HasFlashedMessages(0) {
		d["err"] = self.errorMsg
	}
	if self.HasFlashedMessages(1) {
		d["msg"] = self.successMsg
	}
	d["ctx"] = self
	d["vars"] = self.tplData
	self.Response.SetContentType("html")
	tpl.Execute(self.Response.ResponseWriter, d)
}

//渲染文本
func (self *Handler) RenderText(content string) {
	self.FlushSession()
	self.Response.SetContentType("text")
	fmt.Fprint(self.Response, content)
}

//渲染Json
func (self *Handler) RenderJson(object interface{}) {
	self.FlushSession()
	self.Response.SetContentType("json")
	b, _ := json.Marshal(object)
	fmt.Fprint(self.Response, string(b))
}

//设置cookie
func (self *Handler) SetCookie(key, value string, age int) {
	cookie := http.Cookie{Name: key, Value: value, Path: "/"}
	if age != 0 {
		cookie.MaxAge = age
	}
	http.SetCookie(self.Response.ResponseWriter, &cookie)
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

//刷错误消息
func (self *Handler) FlashError(key, msg string) {
	self.errorMsg[key] = msg
}

//刷成功消息
func (self *Handler) FlashSuccess(key, msg string) {
	self.successMsg[key] = msg
}

//判断是否有消息被刷
func (self *Handler) HasFlashedMessages(mtype int) bool {
	if mtype == 0 {
		return len(self.errorMsg) > 0
	} else {
		return len(self.successMsg) > 0
	}
}
