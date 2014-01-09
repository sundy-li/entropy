package entropy

import (
	"fmt"
	"html/template"
)

type IField interface {
	Label(value string, attrs []string) template.HTML
	Render(attrs []string) template.HTML
	Validate() bool
	GetName() string
	GetValue() string
	SetValue(value string)
	IsName(name string) bool
	HasErrors() bool
	GetErrors() []string
	AddError(err string)
}

type BaseField struct {
	name       string
	label      string
	value      string
	errors     []string
	validators []IValidator
}

func (field *BaseField) Label(value string, attrs []string) template.HTML {
	attrsStr, vstr := "", ""
	if len(attrs) > 0 {
		for _, attr := range attrs {
			attrsStr += " " + template.HTMLEscapeString(attr)
		}
	}
	if len(value) != 0 {
		vstr = value
	} else {
		vstr = field.label
	}
	return template.HTML(fmt.Sprintf("<label for=\"%s\" %s>%s</label>", field.name, attrsStr, vstr))
}

func (field *BaseField) HasErrors() bool {
	return len(field.errors) > 0
}

func (field *BaseField) Render(attrs []string) template.HTML {
	return template.HTML("")
}

func (field *BaseField) GetName() string {
	return field.name
}

func (field *BaseField) AddError(err string) {
	field.errors = append(field.errors, err)
}

func (field *BaseField) GetErrors() []string {
	return field.errors
}

func (field *BaseField) Validate() bool {
	// 如果有Required并且输入为空,不再进行其他检查
	for _, validator := range field.validators {
		if _, ok := validator.(Required); ok {
			if ok, message := validator.Verify(field.GetValue()); !ok {
				field.errors = append(field.errors, field.label+message)
				return false
			}
		}
	}

	result := true

	for _, validator := range field.validators {
		if ok, message := validator.Verify(field.GetValue()); !ok {
			result = false
			field.errors = append(field.errors, message)
		}
	}

	return result
}

func (field *BaseField) GetValue() string {
	return field.value
}

func (field *BaseField) SetValue(value string) {
	field.value = value
}

func (field *BaseField) IsName(name string) bool {
	return field.name == name
}

type TextField struct {
	BaseField
}

func (field *TextField) Render(attrs []string) template.HTML {
	attrsStr := ""
	if len(attrs) > 0 {
		for _, attr := range attrs {
			attrsStr += " " + template.HTMLEscapeString(attr)
		}
	}
	field.value = template.HTMLEscapeString(field.value)
	return template.HTML(fmt.Sprintf(`<input type="text" value="%s" name=%q id=%q%s>`, field.value, field.name, field.name, attrsStr))
}

func NewTextField(name string, label string, value string, validators ...IValidator) *TextField {
	field := TextField{}
	field.name = name
	field.label = label
	field.value = value
	field.validators = validators

	return &field
}

type PasswordField struct {
	BaseField
}

func (field *PasswordField) Render(attrs []string) template.HTML {
	attrsStr := ""
	if len(attrs) > 0 {
		for _, attr := range attrs {
			attrsStr += " " + attr
		}
	}
	return template.HTML(fmt.Sprintf(`<input type="password" name=%q id=%q%s>`, field.name, field.name, attrsStr))
}

func NewPasswordField(name string, label string, validators ...IValidator) *PasswordField {
	field := PasswordField{}
	field.name = name
	field.label = label
	field.validators = validators

	return &field
}

type TextArea struct {
	BaseField
}

func (field *TextArea) Render(attrs []string) template.HTML {
	attrsStr := ""
	if len(attrs) > 0 {
		for _, attr := range attrs {
			attrsStr += " " + attr
		}
	}

	return template.HTML(fmt.Sprintf(`<textarea id=%q name=%q%s>%s</textarea>`, field.name, field.name, attrsStr, field.value))
}

func NewTextArea(name string, label string, value string, validators ...IValidator) *TextArea {
	field := TextArea{}
	field.name = name
	field.label = label
	field.value = value
	field.validators = validators

	return &field
}

type Choice struct {
	Value string
	Label string
}

type SelectField struct {
	BaseField
	Choices []Choice
}

func (field *SelectField) Render(attrs []string) template.HTML {
	attrsStr := ""
	if len(attrs) > 0 {
		for _, attr := range attrs {
			attrsStr += " " + attr
		}
	}
	options := ""
	for _, choice := range field.Choices {
		selected := ""
		if choice.Value == field.value {
			selected = " selected"
		}
		options += fmt.Sprintf(`<option value=%q%s>%s</option>`, choice.Value, selected, choice.Label)
	}

	return template.HTML(fmt.Sprintf(`<select id=%q name=%q%s>%s</select>`, field.name, field.name, attrsStr, options))
}

func NewSelectField(name string, label string, choices []Choice, defaultValue string, validators ...IValidator) *SelectField {
	field := SelectField{}
	field.name = name
	field.label = label
	field.value = defaultValue
	field.Choices = choices

	return &field
}

type HiddenField struct {
	BaseField
}

func (field *HiddenField) Render(attrs []string) template.HTML {
	return template.HTML(fmt.Sprintf(`<input type="hidden" value=%q name=%q id=%q>`, field.value, field.name, field.name))
}

func NewHiddenField(name string, value string) *HiddenField {
	field := HiddenField{}
	field.name = name
	field.value = value

	return &field
}
