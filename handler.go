package entropy

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"time"
)

//处理器接口
type IHandler interface {
	Initialize(name string, cname string, rw http.ResponseWriter, req *http.Request, app *Application)
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
	Name        string
	CName       string
	startTime   time.Time
	Response    Response
	Request     *http.Request
	Session     *Session
	Application *Application
	Messages    map[string]string
	TplData     map[string]interface{}
	Form        *Form
}

//初始化请求处理器
func (self *Handler) Initialize(name string, cname string, rw http.ResponseWriter, req *http.Request, app *Application) {
	self.startTime = time.Now()
	self.Name = name
	self.CName = cname
	self.Request = req
	self.Response = Response{rw}
	self.Application = app
	self.Messages = make(map[string]string)
	self.TplData = make(map[string]interface{})
	if self.Application.Session == nil {
		self.Session = &Session{
			store: NewCookieSession(app.Setting.SessionCookieName, self),
		}
	}
	self.RestoreSession()
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

func (self *Handler) GetQuery(paramName string, defaultValue string) string {
	if param, ok := self.Request.Form[paramName]; ok {
		return param[0]
	} else {
		return defaultValue
	}
}

func (self *Handler) GetQueries(paramName string) []string {
	if param, ok := self.Request.Form[paramName]; ok {
		return param
	} else {
		return nil
	}
}

func (self *Handler) RestoreSession() {
	_tmp, err := self.GetSecureCookie(self.Application.Setting.FlashCookieName)
	if err == nil {
		if err := json.Unmarshal([]byte(_tmp), &self.Messages); err != nil {
			log.Println(err)
		}
	}
	sessionId, err := self.GetCookie(self.Application.Setting.SessionIdCookieName)
	if err != nil {
		log.Println(err.Error())
		sessionId = base64.StdEncoding.EncodeToString([]byte(time.Now().Format(time.RFC3339)))[22:30] + randString(8)
		self.SetCookie(self.Application.Setting.SessionIdCookieName, sessionId, 20)
	}
	log.Println(sessionId)
	self.Session.Restore(sessionId)
}

func (self *Handler) FlushSession() {
	_tmp, err := json.Marshal(self.Messages)
	if err == nil {
		self.SetSecureCookie(self.Application.Setting.FlashCookieName, string(_tmp), 2)
	}
	self.Session.Flush()
}

func (self *Handler) GetStartTime() time.Time {
	return self.startTime
}

//跳转
func (self *Handler) Redirect(url string) {
	self.FlushSession()
	redirectScripts := fmt.Sprintf(
		`<script language="javascript">
		function redirect() {
			location.href="%s";
		}
		setTimeout(redirect,1);
		</script>`, url)
	self.Response.Write([]byte(redirectScripts))
}

//reverse
func (self *Handler) Reverse(name string, arg ...interface{}) string {
	if spec, ok := self.Application.NamedHandlers[name]; ok {
		url, err := spec.UrlSetParams(arg...)
		if err != nil {
			return err.Error()
		} else {
			return url[1:]
		}
	}
	return fmt.Sprintf("处理器 %s 没有找到", name)
}

//赋值到模板变量中
func (self *Handler) Assign(name string, value interface{}) {
	self.TplData[name] = value
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

func (self *Handler) GenerateXsrfHtml() template.HTML {
	xsrfStr := base64.StdEncoding.EncodeToString([]byte(time.Now().Format(time.RFC3339)))[22:30] + randString(8)
	self.SetSecureCookie(XSRF, xsrfStr, 600)
	return template.HTML(fmt.Sprintf(`<input type="hidden" value="%s" name=%q id=%q>`, xsrfStr, XSRF, XSRF))
}

//渲染模板
func (self *Handler) Render(tplPath string) {
	tpl := self.Application.TplEngine.Lookup(tplPath)
	if tpl == nil {
		panic("没有找到指定的模板！")
	}
	tpl.Funcs(self.Application.TplFuncs)
	self.FlushSession()
	self.Response.SetContentType("html")
	tpl.Execute(self.Response.ResponseWriter, self)
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

//刷消息
func (self *Handler) Flash(key, msg string) {
	self.Messages[key] = msg
}

//判断是否有消息被刷
func (self *Handler) HasFlashedMessages() bool {
	return len(self.Messages) > 0
}
