package entropy

type Handler interface{}

//如果返回的布尔值为True,则继续运行,否则跳出,执行Result
type Filter func(*Context) (bool, Result)
