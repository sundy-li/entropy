package validators

import (
	"regexp"
)

type Required struct {
}

func (v Required) Verify(value string) (bool, string) {
	if value == "" {
		return false, "不能为空!"
	}

	return true, ""
}

type Regexp struct {
	Expr    string
	Message string
}

func (v Regexp) Verify(value string) (bool, string) {
	reg, err := regexp.Compile(v.Expr)
	if err != nil {
		panic(err)
	}

	if reg.MatchString(value) {
		return true, ""
	}

	return false, v.Message
}

type Email struct {
}

func (v Email) Verify(value string) (bool, string) {
	tmp := Regexp{Expr: `^.+@[^.].*\.[a-z]{2,10}$`, Message: "无效的电子邮件地址"}

	return tmp.Verify(value)
}

type URL struct {
}

func (v URL) Verify(value string) (bool, string) {
	tmp := Regexp{Expr: `^(http|https)?://([^/:]+|([0-9]{1,3}\.){3}[0-9]{1,3})(:[0-9]+)?(\/.*)?$`, Message: "无效的URL"}

	return tmp.Verify(value)
}

type Int struct {
}

func (v Int) Verify(value string) (bool, string) {
	tmp := Regexp{Expr: `^[0-9]*$`, Message: "无效的整数"}

	return tmp.Verify(value)
}
