package entropy

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
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

//To transform /:path/:action/:id into :/(\w+)/(\w+)/(\w+)
func (self *URLSpec) Url2Regexp() (exp *regexp.Regexp, err error) {
	paramRegexp, _ := regexp.Compile(`:\w+`)
	tmp := paramRegexp.ReplaceAllString(self.Pattern, `(\w+)`)
	exp, err = regexp.Compile(tmp)
	return
}

//Replace the vars into url ，url：/home/:name/:id/:newId, vars ...interface{}
func (self *URLSpec) UrlSetParams(args ...interface{}) (url string, err error) {
	url = self.Pattern
	if strings.HasSuffix(self.Pattern, "$") {
		self.Pattern = strings.TrimSuffix(self.Pattern, "$")
	}
	exp, _ := regexp.Compile(`(:\w+)`)
	matched := exp.FindAllString(self.Pattern, -1)
	if len(matched) != len(args) {
		url = ""
		err = errors.New(fmt.Sprintf("The handler requires %d args , but you provides %d args", len(matched), len(args)))
		return
	} else {
		for index, arg := range args {
			url = strings.Replace(url, matched[index], arg.(string), -1)
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
