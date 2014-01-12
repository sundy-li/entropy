package entropy

import (
	"html/template"
	"net/http"
	"reflect"
	"strings"
)

type Form struct {
	fields map[string]IField
}

func NewForm(formParse interface{}) *Form {
	form := Form{}
	form.fields = make(map[string]IField)
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
	for name, filed := range form.fields {
		filed.SetValue(strings.TrimSpace(r.FormValue(name)))
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
		if !field.Validate() {
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
	field := form.fields[name]
	return field.Render(class, attrs)
}

func (form *Form) Value(name string) string {
	return form.fields[name].GetValue()
}

func (form *Form) SetValue(name, value string) {
	form.fields[name].SetValue(value)
}

func (form *Form) AddError(name, err string) {
	field := form.fields[name]
	field.AddError(err)
}

func (form *Form) Errors() []string {
	var errors []string
	for _, field := range form.fields {
		for _, err := range field.GetErrors() {
			errors = append(errors, err)
		}
	}
	return errors
}
