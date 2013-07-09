package entropy

import (
	"fmt"
	"runtime"
	"path/filepath"
)

//收集错误信息
func MakeStack() []string {
	var stack = make([]string,0)
	ps := make([]uintptr, 20)
	count := runtime.Callers(0, ps)
	for i := 0; i < count; i++ {
		file, line := runtime.FuncForPC(ps[i]).FileLine(ps[i])
		if filepath.Ext(file) == ".go" {
			stack = append(stack, fmt.Sprintf("File:%s Line:%d", file, line))
		} else {
			continue
		}
	}
	return stack
}
