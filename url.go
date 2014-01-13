package entropy

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

//一个URL标识
type URLSpec struct {
	//开发者设置的路径
	Pattern string
	//该路径生成的正则表达式
	Regex *regexp.Regexp
	//处理该路径请求的处理器反射值
	Handler reflect.Value
	//英文名称
	eName string
	//中文名称
	cName string
}

//URLSpec的构造函数
func NewURLSpec(pattern string, handler reflect.Value, ename string, cname string) *URLSpec {
	spec := &URLSpec{Pattern: pattern, Handler: handler, eName: ename, cName: cname}
	var err error
	//将原始路径转为正则表达式
	spec.Regex, err = spec.Url2Regexp()
	if err != nil {
		panic(err)
	}
	return spec
}

//将 /:path/:action/:id 这样的路径转为正则表达式 :/(\w+)/(\w+)/(\w+)
func (self *URLSpec) Url2Regexp() (exp *regexp.Regexp, err error) {
	paramRegexp, _ := regexp.Compile(`:\w+`)
	tmp := paramRegexp.ReplaceAllString(self.Pattern, `(\w+)`)
	exp, err = regexp.Compile(tmp)
	return
}

//将用户实际访问路径中提供的变量值替换到原始路径中 ，url：/home/:name/:id/:newId, vars ...interface{}
func (self *URLSpec) UrlSetParams(args ...interface{}) (url string, err error) {
	url = self.Pattern
	if strings.HasSuffix(url, "$") {
		url = strings.TrimSuffix(url, "$")
	}
	exp, _ := regexp.Compile(`(:\w+)`)
	matched := exp.FindAllString(self.Pattern, -1)
	if len(matched) != len(args) {
		err = errors.New(fmt.Sprintf("严重错误:该处理器需要 %d 个参数 , 但是提供了 %d 个参数.", len(matched), len(args)))
		return
	} else {
		for index, arg := range args {
			if s, ok := arg.(string); ok {
				url = strings.Replace(url, matched[index], s, -1)
			} else if i, ok := arg.(int64); ok {
				url = strings.Replace(url, matched[index], strconv.FormatInt(i, 10), -1)
			} else {
				err = errors.New(fmt.Sprintf("参数%v须为string或int64类型", arg))
				return
			}

		}
		return url, nil
	}
}

//To transform /home/aaa/bbb (/home/:p1/:p2) into map[string][]string   :(\w+):
func (self *URLSpec) ParseUrlParams(url string) (args map[string][]string) {
	paramValues := self.Regex.FindStringSubmatch(url)[1:]
	args = make(map[string][]string, 0)
	paramNameRegexp, _ := regexp.Compile(`(:\w+)`)
	parttens := strings.Split(self.Pattern, "/")
	var index int = 0
	for _, p := range parttens {
		if strings.HasPrefix(p, ":") {
			pName := paramNameRegexp.FindStringSubmatch(p)[1]
			args[strings.TrimPrefix(pName, ":")] = []string{paramValues[index]}
			index++
		}
	}
	return
}
