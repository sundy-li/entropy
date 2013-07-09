package entropy

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type URLSpec struct {
	Pattern string
	Regex   *regexp.Regexp
	Handler reflect.Value
	eName   string
	cName   string
}

func NewURLSpec(pattern string, handler reflect.Value, ename string, cname string) *URLSpec {
	spec := &URLSpec{Pattern: pattern, Handler: handler, eName: ename, cName: cname}
	var err error
	spec.Regex, err = spec.Url2Regexp()
	if err != nil {
		panic(err)
	}
	return spec
}

//To transform /:path:path/:str:action/int:id into :/(.*)/(\w+)/(\d+)
func (self *URLSpec) Url2Regexp() (exp *regexp.Regexp, err error) {
	intRegexp, _ := regexp.Compile(`(:int:\w+)`)
	strRegexp, _ := regexp.Compile(`(:str:\w+)`)
	pathRegexp, _ := regexp.Compile(`(:path:\w+)`)
	tmp := intRegexp.ReplaceAllString(self.Pattern, `(\d+)`)
	tmp = strRegexp.ReplaceAllString(tmp, `(\w+)`)
	tmp = pathRegexp.ReplaceAllString(tmp, `(.*)`)
	exp, err = regexp.Compile(tmp)
	return
}

//Replace the vars into url ，url：/hello/:str:name/:int:id/:int:newId, vars ...interface{}
func (self *URLSpec) UrlSetParams(args ...interface{}) (url string, err error) {
	if strings.HasSuffix(self.Pattern, "$") {
		self.Pattern = strings.TrimSuffix(self.Pattern, "$")
	}
	exp, _ := regexp.Compile(`(:\w+:\w+)`)
	matched := exp.FindAllStringSubmatch(self.Pattern, -1) //[[:str:name :str:name] [:int:id :int:id] [:int:newId :int:newId]]
	if len(matched) != len(args) {
		url = ""
		err = errors.New(fmt.Sprintf("The handler requires %d args , but you provides %d args", len(matched), len(args)))
		return
	} else {
		for index, arg := range args {
			if intArg, ok := arg.(int); ok {
				url = strings.Replace(self.Pattern, matched[index][0], strconv.Itoa(intArg), 1)
			} else {
				url = strings.Replace(self.Pattern, matched[index][0], arg.(string), 1)
			}
		}
		return url, nil
	}
}

//To transform /home/aaa/bbb (/home/:str:p1/:str:p2) into []reflect.Value   :(\w+):
func (self *URLSpec) ParseUrlParams(url string) (args []reflect.Value) {
	paramValues := self.Regex.FindStringSubmatch(url)[1:]
	args = make([]reflect.Value, 0)
	paramTypeRegexp, _ := regexp.Compile(`:(\w+):`)
	parttens := strings.Split(self.Pattern, "/")
	var index int = 0
	for _, p := range parttens {
		if strings.HasPrefix(p, ":") {
			pType := paramTypeRegexp.FindStringSubmatch(p)[1]
			if pType == "int" {
				arg, _ := strconv.ParseInt(paramValues[index], 10, 32)
				args = append(args, reflect.ValueOf(int(arg)))
			} else {
				args = append(args, reflect.ValueOf(paramValues[index]))
			}
			index++
		}
	}
	return
}
