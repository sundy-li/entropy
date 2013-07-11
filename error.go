package entropy

import (
	"html/template"
	"net/http"
)

var (
	ErrHandlers = make(map[int]http.HandlerFunc)
)

func init() {
	ErrHandlers[404] = NotFoundErrorHandler

}

//404默认处理函数
func NotFoundErrorHandler(rw http.ResponseWriter, req *http.Request) {
	t, err := template.New("NotFound").Parse(errorTpl)
	if err != nil {
		panic(err)
	}
	d := make(map[string]interface{})
	d["Code"] = 404
	d["Title"] = "页面没有找到 = =#"
	d["Messages"] = []string{"该页面可能去打酱油了，请稍候再试！", "如果这已经是第二次出现，请检查输入的链接是否正确……", "如果均以确认，请参照第一条……"}
	d["Version"] = EntropyVersion
	t.Execute(rw, d)
}

//500错误默认处理函数
func InternalServerErrorHandler(rw http.ResponseWriter, req *http.Request, code int, err error, debug bool) {
	t, _ := template.New("Error").Parse(errorTpl)
	d := make(map[string]interface{})
	d["Code"] = code
	d["Version"] = EntropyVersion
	d["Title"] = err.Error()
	if debug {
		d["Messages"] = MakeStack()
	} else {
		d["Messages"] = []string{"很抱歉，应用程序发生了错误！"}
	}
	t.Execute(rw, d)
}

var errorTpl = `
<!doctype html>
<html lang="en-US">
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
<title>{{ .Title }} {{ .Version }}</title>
<style>
*{
  margin: 0;
  padding: 0;
  border: 0;
  font-size: 100%;
  font: inherit;
  vertical-align: baseline;
  outline: none;
}
html { height: 100%; }
header { display: block; }
ol, ul { list-style: none; }
a { text-decoration: none }
a:hover { text-decoration: underline }
body {
    background: #dfdfdf;
    font-family: Helvetica, Arial, sans-serif;
    overflow: hidden;
	font-size:62.5%;
}
.clear { clear: both }
.clear:before,
.container:after {
    content: "";
    display: table;
}
.clear:after { clear: both }
.right { float: right }
#main {
    position: relative;
    width: 600px;
    margin: 0 auto;
    padding-top: 8%;
}
#main #header h1 {
    position: relative;
    display: block;
    font: 72px 'Microsoft YaHei', Arial, sans-serif;
    color: #0061a5;
    text-shadow: 2px 2px #f7f7f7;
    text-align: center;
}
#main #header h1 span.sub {
    position: relative;
    font-size: 21px;
    top: -20px;
    padding: 0 10px;
    font-style: italic;
}
#main #header h1 span.icon {
    position: relative;
    display: inline-block;
    top: -6px;
    margin: 0 10px 5px 0;
    background: #0061a5;
    width: 50px;
    height: 50px;
    -moz-box-shadow: 1px 2px white;
    -webkit-box-shadow: 1px 2px white;
    box-shadow: 1px 2px white;
    color: #dfdfdf;
    font-size: 46px;
    line-height: 48px;
    font-weight: bold;
    text-align: center;
    text-shadow: 0 0;
}
#main #content {
    position: relative;
    width: 600px;
    background: white;
    -moz-box-shadow: 0 0 0 3px #ededed inset, 0 0 0 1px #a2a2a2, 0 0 20px rgba(0,0,0,.15);
    -webkit-box-shadow: 0 0 0 3px #ededed inset, 0 0 0 1px #a2a2a2, 0 0 20px rgba(0,0,0,.15);
    box-shadow: 0 0 0 3px #ededed inset, 0 0 0 1px #a2a2a2, 0 0 20px rgba(0,0,0,.15);
    z-index: 5;
}
#main #content h2 {
    background-position: bottom;
    padding: 12px 0 22px 0;
    font: 20px 'Microsoft YaHei', Arial, sans-serif;
    color: #8e8e8e;
    text-align: center;
}
#main #content p {
    position: relative;
    padding: 20px;
    font-size: 13px;
    line-height: 1.3em;
    color: #b5b5b5;
}
#main #content .utilities { padding: 20px }
#main #content .utilities .button {
    display: inline-block;
    height: 34px;
    margin: 0 0 0 6px;
    padding: 0 18px;
    background: #006db0;
    background-image: linear-gradient(bottom, #0062a6 0%, #0079bb 100%);
    background-image: -o-linear-gradient(bottom, #0062a6 0%, #0079bb 100%);
    background-image: -moz-linear-gradient(bottom, #0062a6 0%, #0079bb 100%);
    background-image: -webkit-linear-gradient(bottom, #0062a6 0%, #0079bb 100%);
    background-image: -ms-linear-gradient(bottom, #0062a6 0%, #0079bb 100%);
    -moz-box-shadow: 0 0 0 1px #003255, 0 1px 3px rgba(0, 50, 85, 0.5), 0 1px #00acd8 inset;
    -webkit-box-shadow: 0 0 0 1px #003255, 0 1px 3px rgba(0, 50, 85, 0.5), 0 1px #00acd8 inset;
    box-shadow: 0 0 0 1px #003255, 0 1px 3px rgba(0, 50, 85, 0.5), 0 1px #00acd8 inset;
    font-size: 14px;
    line-height: 34px;
    color: white;
    font-weight: bold;
    text-shadow: 0 -1px #00385a;
    text-decoration: none;
}
#main #content .utilities .button:hover {
    background: #0081c6;
    background-image: linear-gradient(bottom, #006fbb 0%, #008dce 100%);
    background-image: -o-linear-gradient(bottom, #006fbb 0%, #008dce 100%);
    background-image: -moz-linear-gradient(bottom, #006fbb 0%, #008dce 100%);
    background-image: -webkit-linear-gradient(bottom, #006fbb 0%, #008dce 100%);
    background-image: -ms-linear-gradient(bottom, #006fbb 0%, #008dce 100%);
    -moz-box-shadow: 0 0 0 1px #003255, 0 1px 3px rgba(0, 50, 85, 0.5), 0 1px #00c1e4 inset;
    -webkit-box-shadow: 0 0 0 1px #003255, 0 1px 3px rgba(0, 50, 85, 0.5), 0 1px #00c1e4 inset;
    box-shadow: 0 0 0 1px #003255, 0 1px 3px rgba(0, 50, 85, 0.5), 0 1px #00c1e4 inset;
}
#main #content .utilities .button:active {
    background: #0081c6;
    background-image: linear-gradient(bottom, #008dce 0%, #006fbb 100%);
    background-image: -o-linear-gradient(bottom, #008dce 0%, #006fbb 100%);
    background-image: -moz-linear-gradient(bottom, #008dce 0%, #006fbb 100%);
    background-image: -webkit-linear-gradient(bottom, #008dce 0%, #006fbb 100%);
    background-image: -ms-linear-gradient(bottom, #008dce 0%, #006fbb 100%);
}
#main #content .utilities .button-container .button:focus { color: black }
</style>
<!--[if lt IE 9]>
  <script src="http://html5shiv.googlecode.com/svn/trunk/html5.js"></script>
<![endif]-->
</head>
<body>
  <div id="main">
    <header id="header">
      <h1><span class="icon">!</span>{{ .Code }}<span class="sub">{{ .Title }}</span></h1>
    </header>
    <div id="content">
    <p>
    <ol>
    	{{range .Messages}}
    		<li>{{.}}</li>
    	{{end}}
    </ol>
    </p>
      <div class="utilities">
        <a class="button right" href="javascript:history.go(-1);">返回</a>
        <div class="clear"></div>
      </div>
    </div>
  </div>
</div>
</html>
`
