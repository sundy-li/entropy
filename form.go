package entropy

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type EForm struct {
	fields map[string]IField
	Xsrf   string
}

func NewEForm(fields ...IField) *EForm {
	form := EForm{}
	form.fields = make(map[string]IField)
	for _, field := range fields {
		form.fields[field.GetName()] = field
	}
	return &form
}

func (form *EForm) Validate(r *http.Request) bool {
	result := true
	for name, field := range form.fields {
		field.SetValue(strings.TrimSpace(r.FormValue(name)))
		if !field.Validate() {
			result = false
		}
	}
	return result
}

func (form *EForm) Label(name string, attrs ...string) template.HTML {
	field := form.fields[name]

	return field.Label(attrs)
}

func (form *EForm) XsrfHtml() template.HTML {
	return template.HTML(fmt.Sprintf(`<input type="hidden" value="%s" name=%q id=%q>`, form.Xsrf, "_xsrf_", "_xsrf_"))
}

func (form *EForm) Render(name string, attrs ...string) template.HTML {
	field := form.fields[name]
	return field.Render(attrs)
}

func (form *EForm) Value(name string) string {
	return form.fields[name].GetValue()
}

func (form *EForm) SetValue(name, value string) {
	form.fields[name].SetValue(value)
}

func (form *EForm) AddError(name, err string) {
	field := form.fields[name]
	field.AddError(err)
}

func (form *EForm) Errors() []string {
	var errors []string
	for _, field := range form.fields {
		for _, err := range field.GetErrors() {
			errors = append(errors, err)
		}
	}
	return errors
}
