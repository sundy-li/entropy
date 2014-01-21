package entropy

import (
	"crypto/md5"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

//框架版本号
var EntropyVersion = "Entropy 0.618"

type Application struct {
	//执行程序所在目录
	AppPath    string
	Blueprints map[string]*Blueprint
	//application级别的请求处理器集合
	NamedHandlers map[string]*URLSpec
	//before
	BeforeFilters []Filter
	AfterFilters  []Filter
	//错误处理器集合
	ErrorHandlers map[int]Filter
	//配置
	Setting *Setting
	//模板函数
	TplFuncs map[string]interface{}
	//模板引擎
	TplEngine *template.Template
}

//初始化程序,包括模板函数和引擎的初始化
func (self *Application) Initialize() {
	//template functions
	self.TplFuncs["static"] = func(url string) string {
		filePath := path.Join(self.AppPath, self.Setting.StaticDir, url)
		fi, err := os.Stat(filePath)
		if err != nil {
			return err.Error()
		}
		h := md5.New()
		io.WriteString(h, fi.ModTime().Format(time.RFC3339))
		hash := string(h.Sum(nil))
		//这里hash中的每一项都是2个字节,取前4位,即hash[:2]
		urlFile := fmt.Sprintf("/%s/%s?v=%x", self.Setting.StaticDir, url, hash[:2])
		return urlFile
	}
	self.TplFuncs["url"] = func(name string, arg ...interface{}) string {
		if strings.Contains(name, ".") {
			_tmp := strings.Split(name, ".")
			if spec, ok := self.Blueprints[_tmp[0]].NamedHandlers[_tmp[1]]; ok {
				url, err := spec.UrlSetParams(arg...)
				if err != nil {
					return err.Error()
				} else {
					return self.Blueprints[_tmp[0]].Prefix + url[1:]
				}
			}
		}
		if spec, ok := self.NamedHandlers[name]; ok {
			url, err := spec.UrlSetParams(arg...)
			if err != nil {
				return err.Error()
			} else {
				return url[1:]
			}
		}
		return fmt.Sprintf("处理器 %s 没有找到", name)
	}
	self.TplFuncs["empty"] = func(i interface{}) bool {
		if i == nil {
			return true
		}
		if s, ok := i.(string); ok {
			return s == ""
		}
		if a, ok := i.([]string); ok {
			return len(a) == 0
		}
		return true
	}
	//程序执行时间,返回毫秒
	self.TplFuncs["eslape"] = func(ctx *Context) string {
		return fmt.Sprintf("%f", time.Since(ctx.GetStartTime()).Seconds()*1000)
	}
	self.TplFuncs["str_in_array"] = func(key string, strs []string) bool {
		if len(strs) == 0 {
			return false
		}
		for _, s := range strs {
			if s == key {
				return true
			}
		}
		return false
	}
	self.TplFuncs["xsrf"] = func(ctx *Context) template.HTML {
		return template.HTML(fmt.Sprintf(`<input type="hidden" value="%s" name=%q id=%q>`, ctx.GetXsrf(), XSRF, XSRF))
	}
	//构造模板引擎
	tplBasePath := path.Join(self.AppPath, self.Setting.TemplateDir)
	dir, err := os.Stat(tplBasePath)
	if err != nil {
		panic("模板引擎初始化失败." + err.Error())
	}
	if dir.IsDir() != true {
		panic(dir.Name() + "不是一个目录.")
	}
	self.TplEngine = template.New(tplBasePath).Funcs(self.TplFuncs)
	//遍历所有模板目录下的文件
	filepath.Walk(tplBasePath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && !strings.HasPrefix(filepath.Base(path), ".") {
			b, err := ioutil.ReadFile(path)
			if err != nil {
				panic(err)
			}
			s := string(b)
			tmplName := path[len(tplBasePath)+1:]
			//将\\替换为/,以防输入时转义
			tmplName = strings.Replace(tmplName, "\\", "/", -1)
			tmpl := self.TplEngine.New(tmplName).Funcs(self.TplFuncs)
			_, err = tmpl.Parse(s)
			if err != nil {
				panic(err)
			}
		}
		return nil
	})
}

//添加处理器
func (self *Application) Handle(pattern string, eName string, cName string, handler Handler) {
	if strings.Contains(eName, ".") {
		panic("名字里面带个点是几个意思!?")
	}
	//pattern:/home/str:action/int:id
	if !strings.HasSuffix(pattern, "$") {
		pattern = pattern + "$"
	}
	if !strings.HasPrefix(pattern, "^") {
		pattern = "^" + pattern
	}
	if _, exist := self.NamedHandlers[eName]; exist {
		panic(fmt.Sprintf("已经有一个名叫 %s 的处理器！", eName))
	}
	self.NamedHandlers[eName] = NewURLSpec(pattern, handler, eName, cName)
}

func (self *Application) Before(filter Filter) {
	self.BeforeFilters = append(self.BeforeFilters, filter)
}

func (self *Application) Blueprint(name string, bp *Blueprint) {
	self.Blueprints[name] = bp
}

//捕获http请求
func (self *Application) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx := NewContext(self, req, rw)
	defer func() {
		if err := recover(); err != nil {
			switch err {
			case 404:
				if handler, ok := self.ErrorHandlers[404]; ok {
					handler(ctx)
				}
			default:
				if e, ok := err.(error); ok {
					InternalServerErrorHandler(ctx, 500, e, self.Setting.Debug)
				} else {
					InternalServerErrorHandler(ctx, 500, errors.New(err.(string)), self.Setting.Debug)
				}

			}
		}
	}()
	rw.Header().Set("Server", EntropyVersion)
	//判断请求路径是否包含已经设置的静态路径
	if strings.HasPrefix(req.URL.Path, fmt.Sprintf("/%s", self.Setting.StaticDir)) || req.URL.Path == "/favicon.ico" {
		self.processStaticRequest(ctx)
		return
	}
	//查找相符的请求处理器
	spec, bp := self.findMatchedRequestHandler(req)
	if spec == nil {
		panic(404)
	} else {
		self.processRequestHandler(spec, bp, ctx)
	}
	return
}

//找到符合当前请求路径的处理器
func (self *Application) findMatchedRequestHandler(req *http.Request) (*URLSpec, *Blueprint) {
	for _, bp := range self.Blueprints {
		if strings.HasPrefix(req.URL.Path, bp.Prefix) {
			for _, spec := range bp.NamedHandlers {
				if spec.Regex.MatchString(strings.TrimPrefix(req.URL.Path, bp.Prefix)) {
					return spec, bp
				}
			}
		}
	}
	for _, spec := range self.NamedHandlers {
		requestUrl := req.URL.Path
		if spec.Regex.MatchString(requestUrl) {
			return spec, nil
		}
	}
	return nil, nil
}

//处理请求
func (self *Application) processRequestHandler(spec *URLSpec, bp *Blueprint, ctx *Context) {
	//处理request参数
	ctx.Req.ParseForm()
	ctx.Req.ParseMultipartForm(1 << 25) // 32M 1<< 25 /1024/1024
	if !ctx.IsAjax() {
		ctx.generateXsrf()
	}
	ctx.restoreMessages()
	//反射该处理方法
	handler := reflect.TypeOf(spec.Handler)
	//根据该请求的路径,将路径中的参数提取处理
	var params []string
	if bp != nil {
		params = spec.ParseUrlParams(strings.TrimPrefix(ctx.Req.URL.Path, bp.Prefix))
	} else {
		params = spec.ParseUrlParams(ctx.Req.URL.Path)

	}

	//构造路径中的参数
	queryArgs := make([]reflect.Value, 0)
	//如果该方法需要的参数大于或等于1(第一个参数必须为ctx),则把路径中的参数构造好,供调用方法时使用
	if handler.NumIn() >= 1 {
		//从1开始,把ctx过滤
		for i := 1; i < handler.NumIn(); i++ {
			queryArgs = append(queryArgs, reflect.ValueOf(params[i-1]))
		}
	}
	//执行应用级别的before
	for _, before := range self.BeforeFilters {
		before(ctx)
	}
	//如果该处理器位于Blueprint下,还要优先执行Blueprint的before
	if bp != nil {
		for _, before := range bp.BeforeFilters {
			before(ctx)
		}
	}
	//调用方法时需要提供的参数
	args := make([]reflect.Value, 0)
	//第一个参数必须是ctx
	args = append(args, reflect.ValueOf(ctx))
	//后续路径中的参数
	args = append(args, queryArgs...)
	//调用方法,获取第一个返回值 Result
	result := reflect.ValueOf(spec.Handler).Call(args)[0].Interface().(Result)
	if bp != nil {
		for _, after := range bp.AfterFilters {
			after(ctx)
		}
	}
	//执行应用级别的after
	for _, after := range self.AfterFilters {
		after(ctx)
	}
	ctx.flushMessage()
	//调用result的execute方法,进行输出
	result.Execute(ctx.Resp)
}

//处理静态文件
func (self *Application) processStaticRequest(ctx *Context) {
	//e.appPath=x://path_to_app req.Url.Path=/<e.Config.StaticDir>/css/style.css
	//静态文件的硬盘路径
	filePath := path.Join(self.AppPath, ctx.Req.URL.Path)
	_, err := os.Stat(filePath)
	if err != nil {
		//不存在则404错误
		panic(404)
	}
	//直接使用ServeFile方法来处理静态文件
	http.ServeFile(ctx.Resp, ctx.Req, path.Join(self.AppPath, ctx.Req.URL.Path))
}

//运行程序
func (self *Application) Go(host string, port int) {
	addr := fmt.Sprintf("%s:%d", host, port)
	go func() {
		fmt.Println("Server is listening at ", addr)
	}()
	log.Fatalln(http.ListenAndServe(addr, self))
}

//构造Application对象
func NewApplication(filePath string) *Application {
	pwd, _ := os.Getwd()
	application := &Application{
		AppPath:       pwd,
		NamedHandlers: make(map[string]*URLSpec),
		Blueprints:    make(map[string]*Blueprint),
		BeforeFilters: make([]Filter, 0),
		AfterFilters:  make([]Filter, 0),
		ErrorHandlers: ErrHandlers, //定义在error.go中
		Setting:       NewSetting(filePath),
		TplFuncs:      make(map[string]interface{}),
	}
	application.Initialize()
	return application
}
