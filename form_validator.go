package entropy

import (
	"regexp"
)

var (
	XValidators = make(map[string]IValidator)
)

func init() {
	XValidators["required"] = &Required{}
	XValidators["email"] = &Regexp{Expr: `^(\w)+(\.\w+)*@(\w)+((\.\w+)+)$`, Message: "不是一个有效的邮件地址"}
}

type IValidator interface {
	Verify(value string) (bool, string)
}

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
