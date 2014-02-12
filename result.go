package entropy

import (
	"bufio"
	"encoding/json"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
)

//结果接口
type Result interface {
	Execute(io.Writer)
}

//===========================Html结果======================
func NewHtmlResult(ctx *Context, tpl string) *HtmlResult {
	return &HtmlResult{ctx, tpl}
}

type HtmlResult struct {
	Context *Context
	Tpl     string
}

func (self *HtmlResult) Execute(writer io.Writer) {
	tpl := self.Context.App.TplEngine.Lookup(self.Tpl)
	if tpl == nil {
		panic("没有找到指定的模板！" + self.Tpl)
	}
	self.Context.Resp.SetContentType("html")
	tpl.Execute(writer, self.Context)
}

//===========================Html结果 end======================

//===========================Text结果======================
func NewTextResult(ctx *Context, content string) *TextResult {
	return &TextResult{ctx, content}
}

type TextResult struct {
	Context *Context
	Content string
}

func (self *TextResult) Execute(writer io.Writer) {
	self.Context.Resp.SetContentType("text")
	writer.Write([]byte(self.Content))
}

//===========================Text结果 end======================

//===========================Json结果======================
func NewJsonResult(ctx *Context, obj interface{}) *JsonResult {
	return &JsonResult{ctx, obj}
}

type JsonResult struct {
	Context *Context
	Object  interface{}
}

func (self *JsonResult) Execute(writer io.Writer) {
	self.Context.Resp.SetContentType("json")
	b, _ := json.Marshal(self.Object)
	writer.Write(b)
}

//===========================Json结果 end======================

//===========================Redirect结果======================
func NewRedirectResult(ctx *Context, url string, forever bool) *RedirectResult {
	return &RedirectResult{ctx, url, forever}
}

type RedirectResult struct {
	Context *Context
	url     string
	forever bool
}

func (self *RedirectResult) Execute(writer io.Writer) {
	self.Context.Resp.SetHeader("Location", self.url, true)
	//self.Context.Resp.Write([]byte("redirecting"))
	if self.forever {
		self.Context.Resp.WriteHeader(301)
	} else {
		self.Context.Resp.WriteHeader(302)
	}

}

//===========================Redirect结果 end======================

//===========================Image结果======================
func NewImageResult(ctx *Context, img image.Image, imgType int) *ImageResult {
	return &ImageResult{ctx, img, imgType}
}

type ImageResult struct {
	Context   *Context
	img       image.Image
	imageType int
}

func (self *ImageResult) Execute(writer io.Writer) {
	b := bufio.NewWriter(writer)
	switch self.imageType {
	case IMAGEPNG:
		{
			png.Encode(b, self.img)
			self.Context.Resp.SetContentType("png")
		}
	case IMAGEGIF:
		{
			gif.Encode(b, self.img, nil)
			self.Context.Resp.SetContentType("gif")
		}
	case IMAGEJPEG:
		{
			jpeg.Encode(b, self.img, nil)
			self.Context.Resp.SetContentType("jpeg")
		}
	default:
		panic("错误的图片类型!")
	}
	b.Flush()
}

//===========================Image结果 end======================
