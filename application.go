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
	AppPath string
	//所有的请求处理器集合
	NamedHandlers map[string]*URLSpec
	//错误处理器集合
	ErrorHandlers map[int]http.HandlerFunc
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
	//程序执行时间,返回毫秒
	self.TplFuncs["eslape"] = func(handler IHandler) string {
		return fmt.Sprintf("%f", time.Since(handler.GetStartTime()).Seconds()*1000)
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
func (self *Application) AddHandler(pattern string, eName string, cName string, handler IHandler) {
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
	self.NamedHandlers[eName] = NewURLSpec(pattern, reflect.ValueOf(handler), eName, cName)
}

//捕获http请求
func (self *Application) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			switch err {
			case 404:
				if handler, ok := self.ErrorHandlers[404]; ok {
					handler(rw, req)
				}
			default:
				if e, ok := err.(error); ok {
					InternalServerErrorHandler(rw, req, 500, e, self.Setting.Debug)
				} else {
					InternalServerErrorHandler(rw, req, 500, errors.New(err.(string)), self.Setting.Debug)
				}

			}
		}
	}()
	rw.Header().Set("Server", EntropyVersion)
	//判断请求路径是否包含已经设置的静态路径
	if strings.HasPrefix(req.URL.Path, fmt.Sprintf("/%s", self.Setting.StaticDir)) || req.URL.Path == "/favicon.ico" {
		self.processStaticRequest(rw, req)
		return
	}
	//查找相符的请求处理器
	spec := self.findMatchedRequestHandler(req)
	if spec == nil {
		panic(404)
	} else {
		self.processRequestHandler(spec, req, rw)
	}
	return
}

//找到符合当前请求路径的处理器
func (self *Application) findMatchedRequestHandler(req *http.Request) (matchedSpec *URLSpec) {
	for _, spec := range self.NamedHandlers {
		var requestUrl string
		requestUrl = req.URL.Path
		if spec.Regex.MatchString(requestUrl) {
			matchedSpec = spec //最后一个match
		}
	}
	return
}

//处理请求
func (self *Application) processRequestHandler(spec *URLSpec, req *http.Request, rw http.ResponseWriter) {
	//处理器的Initialize方法
	methodInitialize := spec.Handler.MethodByName("Initialize")
	argsInitialize := make([]reflect.Value, 3)
	argsInitialize[0] = reflect.ValueOf(rw)
	argsInitialize[1] = reflect.ValueOf(req)
	argsInitialize[2] = reflect.ValueOf(self)
	methodInitialize.Call(argsInitialize)

	//处理器的Prepare方法
	methodPrepare := spec.Handler.MethodByName("Prepare")
	methodPrepare.Call([]reflect.Value{})
	//处理表单
	req.ParseForm()
	req.ParseMultipartForm(1 << 25) // 32M 1<< 25 /1024/1024
	args := spec.ParseUrlParams(req.URL.Path)
	for name, arg := range args {
		req.Form[name] = arg
	}
	//请求所对应的方法
	method := spec.Handler.MethodByName(strings.Title(strings.ToLower(req.Method)))
	method.Call([]reflect.Value{})

	//处理器的Finish方法
	methodFinish := spec.Handler.MethodByName("Finish")
	methodFinish.Call([]reflect.Value{})
}

//处理静态文件
func (self *Application) processStaticRequest(rw http.ResponseWriter, req *http.Request) {
	//e.appPath=x://path_to_app req.Url.Path=/<e.Config.StaticDir>/css/style.css
	//静态文件的硬盘路径
	filePath := path.Join(self.AppPath, req.URL.Path)
	_, err := os.Stat(filePath)
	if err != nil {
		//不存在则404错误
		panic(404)
	}
	//直接使用ServeFile方法来处理静态文件
	http.ServeFile(rw, req, path.Join(self.AppPath, req.URL.Path))
}

//运行程序
func (self *Application) Go(host string, port int) {
	addr := fmt.Sprintf("%s:%d", host, port)
	go func() {
		fmt.Println("Server is listening : ", addr)
	}()
	log.Fatalln(http.ListenAndServe(addr, self))
}

//构造Application对象
func NewApplication(filePath string) *Application {
	pwd, _ := os.Getwd()
	application := &Application{
		AppPath:       pwd,
		NamedHandlers: make(map[string]*URLSpec),
		ErrorHandlers: ErrHandlers, //定义在error.go中
		Setting:       NewSetting(filePath),
		TplFuncs:      make(map[string]interface{}),
	}
	application.Initialize()
	return application
}
