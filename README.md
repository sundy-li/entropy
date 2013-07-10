Entropy [![Build Status](https://drone.io/github.com/frank418/entropy/status.png)](https://drone.io/github.com/frank418/entropy/latest)
=======
一个简单的go语言实现的web框架

示例
----

```go

package main

import (
	e "github.com/frank418/entropy"
)

type HomeHandler struct {
	e.RequestHandler
}

func (self *HomeHandler) Get() {
	self.Render("home.html")
}

type Error404 struct {
	e.RequestHandler
}

func (self *Error404) Get() {
	panic(404)
}

type Error500 struct {
	e.RequestHandler
}

func (self *Error500) Get() {
	panic("我是故意出错的！")
}

type TemplateData struct {
	e.RequestHandler
}

func (self *TemplateData) Get() {
	self.Assign("value", "我是被赋值过来的变量值！")
	self.Render("template.html")
}

type Flash struct {
	e.RequestHandler
}

func (self *Flash) Get() {
	self.FlashSuccess("我是一个成功的信息！")
	self.FlashError("我是一个错误的信息！")
	self.Render("flash.html")
}

type FlashResult struct {
	e.RequestHandler
}

func (self *FlashResult) Get() {
	self.Render("flashresult.html")
}

func main() {
	app := e.NewApplication("app.conf")
	app.AddHandler("/", "Home", "首页", &HomeHandler{})
	app.AddHandler("/404", "404", "404错误", &Error404{})
	app.AddHandler("/500", "500", "500错误", &Error500{})
	app.AddHandler("/tpl", "tpl", "赋值模板变量", &TemplateData{})
	app.AddHandler("/flash", "flash", "刷俩信息", &Flash{})
	app.AddHandler("/result", "result", "看上面那俩信息去", &FlashResult{})
	app.Go("", 9999)
}


```

`app.conf`
```json

{
	"Debug":true,
	"TemplateDir":"template",
	"StaticDir":"assets"
}

```


`home.html`
```html

<ol>
<li><a href="{{ url "404" }}" target="_blank">猛击这里去看404</a></li>
<li><a href="{{ url "500" }}" target="_blank">猛击这里去看500</a></li>
<li><a href="{{ url "tpl" }}" target="_blank">猛击这里去围观被赋值的模板变量</a></li>
<li><a href="{{ url "flash" }}" target="_blank">Flash俩信息</a></li>
</ol>

```

`template.html`
```html

{{ .vars.value }}

```

`flash.html`
```html

我Flash了一个错误信息和一个成功信息！<a href="{{ url "result" }}">去下一页查看</a>

```

`flashresult.html`
```html

{{range $.ctx.GetFlashedMessagesWithType "error" }}
 我是一个坏消息：{{.}}
{{end}}

{{range $.ctx.GetFlashedMessagesWithType "success" }}
 我是一个好消息：{{.}}
{{end}}

```
