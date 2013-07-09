package entropy

import (
	"compress/gzip"
	"compress/zlib"
	"crypto/md5"
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

//Version of The Framework , just like the Golden Section.
var EntropyVersion = "Entropy 0.618"

type Application struct {
	AppPath       string
	NamedHandlers map[string]*URLSpec
	ErrorHandlers map[int]http.HandlerFunc
	Setting       *Setting
	TplFuncs      map[string]interface{}
	TplEngine     *template.Template
}

func (self *Application) Initialize() {
	//template functions
	self.TplFuncs["static"] = func(url string) string {
		filePath := path.Join(self.AppPath, self.Setting.StaticDir, url)
		fi, err := os.Stat(filePath)
		if err != nil {
			panic(err)
		}
		hash := md5.New().Sum([]byte(fi.ModTime().Format(time.RFC3339)))
		urlFile := fmt.Sprintf("/%s/%s?v=%x", self.Setting.StaticDir, url, hash[:4])
		return urlFile
	}
	self.TplFuncs["url"] = func(name string, arg ...interface{}) string {
		if spec, ok := self.NamedHandlers[name]; ok {
			url, err := spec.UrlSetParams(arg...)
			if err != nil {
				return err.Error()
			} else {
				return url
			}
		}
		return fmt.Sprintf("Hanlder %s Not Found", name)
	}
	//template engine
	tplBasePath := path.Join(self.AppPath, self.Setting.TemplateDir)
	dir, err := os.Stat(tplBasePath)
	if err != nil {
		panic("Template Engine initialize failed." + err.Error())
	}
	if dir.IsDir() != true {
		panic(dir.Name() + "is not a directory.")
	}
	self.TplEngine = template.New(tplBasePath).Funcs(self.TplFuncs)
	//Go through all files in template dir
	filepath.Walk(tplBasePath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && !strings.HasPrefix(filepath.Base(path), ".") {
			b, err := ioutil.ReadFile(path)
			if err != nil {
				panic(err)
			}
			s := string(b)
			tmplName := path[len(tplBasePath)+1:]
			//Replace \\ to /
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

func (self *Application) AddHandler(pattern string, eName string, cName string, handler IHandler) {
	//pattern:/home/str:action/int:id
	if !strings.HasSuffix(pattern, "$") {
		pattern = pattern + "$"
	}
	if _, exist := self.NamedHandlers[eName]; exist {
		panic(fmt.Sprintf("Already had a Handler named %s！", eName))
	}
	self.NamedHandlers[eName] = NewURLSpec(pattern, reflect.ValueOf(handler), eName, cName)
}

func (self *Application) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			switch err {
			case 404:
				if handler, ok := self.ErrorHandlers[404]; ok {
					handler(rw, req)
				}
			default:
				InternalServerErrorHandler(rw, req, 500, err, self.Setting.Debug)
			}
		}
	}()
	rw.Header().Set("Server", EntropyVersion)
	if strings.HasPrefix(req.URL.Path, fmt.Sprintf("/%s", self.Setting.StaticDir)) || req.URL.Path == "/favicon.ico" {
		self.processStaticRequest(rw, req)
		return
	}
	spec := self.findMatchedRequestHandler(req)
	fmt.Printf("%v:%s\n", spec, req.URL.Path)
	if spec == nil {
		panic(404)
	} else {
		self.processRequestHandler(spec, req, rw)
	}
	return
}

//private functions
func (self *Application) findMatchedRequestHandler(req *http.Request) (matchedSpec *URLSpec) {
	for _, spec := range self.NamedHandlers {
		if spec.Regex.MatchString(req.URL.Path) {
			matchedSpec = spec //最后一个match
		}
	}
	return
}

func (self *Application) processRequestHandler(spec *URLSpec, req *http.Request, rw http.ResponseWriter) {
	//prepare
	transformer := rw.(io.Writer)
	if rw.Header().Get("Accept-Encoding") != "" {
		encodings := strings.SplitN(rw.Header().Get("Accept-Encoding"), ",", -1)
		for i, v := range encodings {
			encodings[i] = strings.TrimSpace(v)
		}
		for _, val := range encodings {
			if val == "gzip" {
				rw.Header().Set("Content-Encoding", "gzip")
				transformer, _ = gzip.NewWriterLevel(rw, gzip.BestSpeed)
				break
			} else if val == "deflate" {
				rw.Header().Set("Content-Encoding", "deflate")
				transformer, _ = zlib.NewWriterLevel(rw, zlib.BestSpeed)
				break
			}
		}
	}
	//handler method calls
	//initialize
	methodInitialize := spec.Handler.MethodByName("Initialize")
	argsInitialize := make([]reflect.Value, 2)
	argsInitialize[0] = reflect.ValueOf(rw)
	argsInitialize[1] = reflect.ValueOf(req)
	methodInitialize.Call(argsInitialize)
	//initRequestHandler method
	methodInit := spec.Handler.MethodByName("InitRequestHandler")
	argsInit := make([]reflect.Value, 2)
	argsInit[0] = reflect.ValueOf(self)
	argsInit[1] = reflect.ValueOf(transformer)
	methodInit.Call(argsInit)
	//prepare method
	methodPrepare := spec.Handler.MethodByName("Prepare")
	methodPrepare.Call([]reflect.Value{})
	//request method
	req.ParseForm()
	req.ParseMultipartForm(1 << 25)
	args := spec.ParseUrlParams(req.URL.Path)
	for name, arg := range args {
		req.Form[name] = arg
	}
	method := spec.Handler.MethodByName(strings.Title(strings.ToLower(req.Method)))
	method.Call([]reflect.Value{})
	//finish method
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

func (self *Application) Go(host string, port int) {
	addr := fmt.Sprintf("%s:%d", host, port)
	go func() {
		fmt.Println("Server is listening : %s", addr)
	}()
	log.Fatalln(http.ListenAndServe(addr, self))
}

func NewApplication(filePath string) *Application {
	pwd, _ := os.Getwd()
	application := &Application{
		AppPath:       pwd,
		NamedHandlers: make(map[string]*URLSpec),
		ErrorHandlers: ErrHandlers,
		Setting:       NewSetting(filePath),
		TplFuncs:      make(map[string]interface{}),
	}
	application.Initialize()
	return application
}
