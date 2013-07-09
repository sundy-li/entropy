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
	Debug           bool
	TemplateDir     string
	StaticDir       string
	CookieSecret    string
	FlashCookieName string
}

//Generate a new setting . if provides a setting json file , load & use that ; otherwise , use the default setting
func NewSetting(fileName string) *Setting {
	cPath, _ := os.Getwd()
	filePath := path.Join(cPath, fileName)
	file, err := ioutil.ReadFile(filePath)
	secret := fmt.Sprintf("%x", sha1.New().Sum([]byte(time.Now().Format(time.RFC3339))))
	c := Setting{
		Debug:           true,
		TemplateDir:     "template",
		StaticDir:       "static",
		CookieSecret:    secret[len(secret)-32:],
		FlashCookieName: "msgs",
	}
	log.Println("Loaded default setting")
	if err == nil {
		json.Unmarshal(file, &c)
		log.Println("Loaded user's setting")
	}
	return &c
}
