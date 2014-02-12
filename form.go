package entropy

import (
	"fmt"
	"html/template"
	"net/http"
	"reflect"
	"strings"
)

type Form struct {
	fields map[string]IField
	errors map[string][]string
}

func NewForm(formParse interface{}) *Form {
	form := Form{}
	form.fields = make(map[string]IField)
	form.errors = make(map[string][]string, 0)
	_form := reflect.ValueOf(formParse).Elem()
	for i := 0; i < _form.NumField(); i++ {
		field := _form.Field(i)
		//把非IField排除
		if f, ok := field.Interface().(IField); ok {
			form.fields[f.GetName()] = f
		}
	}
	return &form
}

/*从请求中分析表单
interface{} 返回存有表单值的实例对象
*Form 返回带方法的Form对象
*/
func ParseForm(formParse interface{}, r *http.Request) (interface{}, *Form) {
	form := NewForm(formParse)
	for name, field := range form.fields {
		field.SetValue(strings.TrimSpace(r.FormValue(name)))
	}
	_form := reflect.ValueOf(formParse).Elem()
	for i := 0; i < _form.NumField(); i++ {
		field := _form.Field(i)
		//把非IField排除
		if f, ok := field.Interface().(IField); ok {
			f = form.fields[f.GetName()]
		}
	}
	return _form.Interface(), form
}

func (form *Form) Validate(r *http.Request) bool {
	result := true
	for _, field := range form.fields {
		ret, err := field.Validate()
		if !ret {
			form.errors[field.GetName()] = append(form.errors[field.GetName()], err)
			result = false
		}
	}
	return result
}

func (form *Form) Label(name string, class string, attrs ...string) template.HTML {
	field := form.fields[name]
	return field.Label(class, attrs)
}

func (form *Form) Render(name string, class string, attrs ...string) template.HTML {
	field, ok := form.fields[name]
	if !ok {
		return template.HTML(fmt.Sprintf("没有名为%s的表单项,请检查表单定义", name))
	} else {
		return field.Render(class, attrs)
	}
}

func (form *Form) Value(name string) string {
	return form.fields[name].GetValue()
}

func (form *Form) SetValue(name, value string) {
	form.fields[name].SetValue(value)
}

func (form *Form) AllErrors() []string {
	var errs []string
	for _, list := range form.errors {
		for _, e := range list {
			errs = append(errs, e)
		}
	}
	return errs
}

func (form *Form) Errors() map[string][]string {
	return form.errors
}
