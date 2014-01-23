package entropy

import (
	"fmt"
	"html/template"

	v "github.com/frank418/entropy/validators"
)

type IField interface {
	Label(class string, attrs []string) template.HTML
	Render(class string, attrs []string) template.HTML
	Validate() (bool, string)
	GetName() string
	GetValue() string
	SetValue(value string)
	IsName(name string) bool
}

type BaseField struct {
	name       string
	label      string
	value      string
	validators []IValidator
}

func (field *BaseField) Label(class string, attrs []string) template.HTML {
	attrsStr := ""
	if len(attrs) > 0 {
		for _, attr := range attrs {
			attrsStr += " " + template.HTMLEscapeString(attr)
		}
	}
	return template.HTML(fmt.Sprintf(`<label for="%s" class="%s" %s>%s</label>`, field.name, class, attrsStr, field.label))
}

func (field *BaseField) Render(class string, attrs []string) template.HTML {
	return template.HTML("")
}

func (field *BaseField) GetName() string {
	return field.name
}

func (field *BaseField) Validate() (bool, string) {
	// 如果有Required并且输入为空,不再进行其他检查
	for _, validator := range field.validators {
		if _, ok := validator.(v.Required); ok {
			if ok, message := validator.Verify(field.GetValue()); !ok {
				return false, field.label + message
			}
		}
	}

	result := true
	msg := ""
	for _, validator := range field.validators {
		if ok, message := validator.Verify(field.GetValue()); !ok {
			result = false
			msg = message
		}
	}

	return result, msg
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

func (field *TextField) Render(class string, attrs []string) template.HTML {
	attrsStr := ""
	if len(attrs) > 0 {
		for _, attr := range attrs {
			attrsStr += " " + template.HTMLEscapeString(attr)
		}
	}
	field.value = template.HTMLEscapeString(field.value)
	return template.HTML(fmt.Sprintf(`<input type="text" class="%s" value="%s" name=%q id=%q%s>`, class, field.value, field.name, field.name, attrsStr))
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

func (field *PasswordField) Render(class string, attrs []string) template.HTML {
	attrsStr := ""
	if len(attrs) > 0 {
		for _, attr := range attrs {
			attrsStr += " " + attr
		}
	}
	return template.HTML(fmt.Sprintf(`<input type="password" class="%s" name=%q id=%q%s>`, class, field.name, field.name, attrsStr))
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

func (field *TextArea) Render(class string, attrs []string) template.HTML {
	attrsStr := ""
	if len(attrs) > 0 {
		for _, attr := range attrs {
			attrsStr += " " + attr
		}
	}

	return template.HTML(fmt.Sprintf(`<textarea id="%s" class="%s" name="%s" %s>%s</textarea>`, field.name, class, field.name, attrsStr, field.value))
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

func (field *SelectField) Render(class string, attrs []string) template.HTML {
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
		options += fmt.Sprintf(`<option value="%s" %s>%s</option>`, choice.Value, selected, choice.Label)
	}

	return template.HTML(fmt.Sprintf(`<select id="%s" class="%s" name="%s" %s>%s</select>`, field.name, class, field.name, attrsStr, options))
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

func (field *HiddenField) Render(class string, attrs []string) template.HTML {
	return template.HTML(fmt.Sprintf(`<input type="hidden" class="%s" value=%q name=%q id=%q>`, class, field.value, field.name, field.name))
}

func NewHiddenField(name string, value string) *HiddenField {
	field := HiddenField{}
	field.name = name
	field.value = value

	return &field
}
