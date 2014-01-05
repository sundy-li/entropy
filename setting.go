package entropy

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

type Setting struct {
	Debug             bool
	TemplateDir       string
	StaticDir         string
	Secret            string
	FlashCookieName   string
	SessionCookieName string
}

var (
	globalSetting *Setting
)

//Setting的构造函数，如果全局setting已经存在，则直接返回，即使传入的是新的配置文件，也不进行处理
func NewSetting(fileName string) *Setting {
	if globalSetting != nil {
		return globalSetting
	} else {
		cPath, _ := os.Getwd()
		filePath := path.Join(cPath, fileName)
		file, err := ioutil.ReadFile(filePath)
		secret := fmt.Sprintf("%x", sha1.New().Sum([]byte(time.Now().Format(time.RFC3339))))
		globalSetting := &Setting{
			Debug:             true,
			TemplateDir:       "template",
			StaticDir:         "static",
			Secret:            secret[len(secret)-32:],
			FlashCookieName:   "msgs",
			SessionCookieName: "session",
		}
		log.Println("Loaded default setting")
		if err == nil {
			json.Unmarshal(file, globalSetting)
			log.Println("Loaded user's setting")
		}
		return globalSetting
	}
}
