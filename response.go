package entropy

import (
	"mime"
	"net/http"
	"strings"
)

//自定义响应结构体
type Response struct {
	http.ResponseWriter
}

//设置http头
func (r *Response) SetHeader(key string, value string, unique bool) {
	//如果值必须是唯一的,使用set;否则,使用add
	if unique {
		r.Header().Set(key, value)
	} else {
		r.Header().Add(key, value)
	}
}

//设置ContentType
func (r *Response) SetContentType(ext string) {
	var contentType string
	//判断传入的扩展名是否有.
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	//如果能从系统mime中获取相应的type，则使用系统提供的mimetype；否则的话，将application/ext直接设置为content type
	if mime.TypeByExtension(ext) != "" {
		contentType = mime.TypeByExtension(ext)
	} else {
		contentType = "application/" + strings.TrimPrefix(ext, ".") + ";charset=utf-8"
	}
	r.SetHeader("Content-Type", contentType, true)
}