package entropy

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
)

type Setting struct {
	Debug             bool
	TemplateDir       string
	StaticDir         string
	Secret            string
	FlashCookieName   string
	SessionCookieName string
	Xsrf              bool
	XsrfCookie        string
	CurrentUser       string
	Capt              string
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
		//secret := fmt.Sprintf("%x", sha1.New().Sum([]byte(time.Now().Format(time.RFC3339))))
		globalSetting := &Setting{
			Debug:             true,
			TemplateDir:       "template",
			StaticDir:         "static",
			FlashCookieName:   "entropy_msg",
			SessionCookieName: "entropy_session",
			Xsrf:              true,
			XsrfCookie:        "entropy_csrf",
			CurrentUser:       "entropy_user",
			Capt:              "entropy_capt",
		}
		log.Println("Loaded default setting")
		if err == nil {
			err = json.Unmarshal(file, globalSetting)
			if err != nil {
				panic(err.Error())
			} else {
				log.Println("Loaded user's setting")
			}

		}
		if globalSetting.Secret == "" {
			log.Println(globalSetting)
			panic("必须提供一个密匙！Secret!")
		}
		return globalSetting
	}
}
