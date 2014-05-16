package entropy

import (
	"fmt"
	"path/filepath"
	"runtime"
)

//收集错误信息
func MakeStack() []string {
	var stack = make([]string, 0)
	ps := make([]uintptr, 300)
	count := runtime.Callers(0, ps)
	for i := 0; i < count; i++ {
		file, line := runtime.FuncForPC(ps[i]).FileLine(ps[i])
		//仅收集扩展名为go的文件信息
		if filepath.Ext(file) == ".go" {
			stack = append(stack, fmt.Sprintf("File:%s Line:%d", file, line))
		} else {
			continue
		}
	}
	return stack
}
