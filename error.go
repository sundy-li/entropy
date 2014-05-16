package entropy

import (
	"html/template"
)

var (
	ErrHandlers = make(map[int]Filter)
)

func init() {
	ErrHandlers[404] = NotFoundErrorHandler

}

//404默认处理函数
func NotFoundErrorHandler(ctx *Context) (b bool, r Result) {
	b = true
	r = nil
	ctx.Resp.WriteHeader(404)
	t, err := template.New("NotFound").Parse(errorTpl)
	if err != nil {
		panic(err)
	}
	d := make(map[string]interface{})
	d["Code"] = 404
	d["Title"] = "页面没有找到 = =#"
	d["Messages"] = []string{"该页面可能去打酱油了，请稍候再试！", "如果这已经是第二次出现，请检查输入的链接是否正确……", "如果均已确认，请参照第一条……"}
	d["Version"] = EntropyVersion
	t.Execute(ctx.Resp, d)
	return
}

//500错误默认处理函数
func InternalServerErrorHandler(ctx *Context, code int, err error, debug bool) {
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
	t.Execute(ctx.Resp, d)
}

var errorTpl = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <!--[if IE]><meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1"><![endif]-->
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Entropy | Error {{.Code}}</title>
    <style>
    body {
        font-family: "Monaco", sans-serif, "Helvetica Neue", Helvetica, Arial, sans-serif;
        font-size: 12px;
        line-height: 1.428571429;
        color: #949494;
        background-color: #ffffff;
    }
    .page-container {
        position: relative;
    }
    .page-container:before,
    .page-container:after {
        content: " ";
        display: table;
    }
    .page-container:after {
        clear: both;
    }
    .page-container .main-content {
        position: relative;
        float: left;
        width: 100%;
        padding: 20px;
        z-index: 2;
        background: #ffffff;
        -webkit-box-sizing: border-box;
        -moz-box-sizing: border-box;
        box-sizing: border-box;
        }.page-error {
            color: #303641;
            text-align: center;
        }

        .page-error .error-text {
            padding-bottom: 25px;
            font-size: 16px;
        }

        .page-error .error-text h2 {
            font-size: 45px;
        }

        .page-error .error-text p {
            font-size: 22px;
        }

        .page-error .error-text + hr {
            margin-bottom: 50px;
        }

        .page-error .input-group {
            width: 250px;
            margin: 0 auto;
        }

        </style>

    </head>
    <body class="page-body">

        <div class="page-container">
            <div class="main-content">
                <div class="page-error">

                    <div class="error-text">
                        <h2>{{.Title}}</h2>
                            {{range .Messages}}
                                <p>{{.}}</p>
                            {{end}}
                    </div>

                    <hr />
                </div>

            </div>
        </div>

    </body>
    </html>

`
