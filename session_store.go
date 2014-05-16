package entropy

import (
	"encoding/json"
	"log"
)

var (
	//全局session存储
	globalSessionStore ISessionStore
)

//CookieSession构造函数,使用全局单例,因为一个应用不可能要求session处于多种存储状态
func NewCookieSession(sessionKey string, ctx *Context) ISessionStore {
	if globalSessionStore == nil {
		globalSessionStore = &CookieSession{
			SessionData: make(map[string]interface{}),
			sessionKey:  sessionKey,
			ctx:         ctx,
		}
	}
	return globalSessionStore
}

//CookieSession 结构体
type CookieSession struct {
	SessionData map[string]interface{}
	sessionKey  string
	ctx         *Context
}

//恢复cookie中的数据到SessionData中
func (self *CookieSession) Restore() {
	sessionStr, err := self.ctx.SecureCookie(self.sessionKey)
	if err != nil {
		self.SessionData = make(map[string]interface{})
		return
	}

	err = json.Unmarshal([]byte(sessionStr), &self.SessionData)
	if err != nil {
		log.Println("CookieSession Restore", sessionStr)
		self.SessionData = make(map[string]interface{})
		return
	}
}

//将SessionData中的数据写入到cookie中
func (self *CookieSession) Flush(age int) {
	if age < 0 {
		self.ctx.SetSecureCookie(self.sessionKey, "", age)
		return
	}
	sessionByte, err := json.Marshal(self.SessionData)
	if err != nil {
		log.Printf("marshal %#v %s", self.SessionData, err)
	}
	log.Printf("%#v", self.SessionData)
	for key, _ := range self.SessionData {
		delete(self.SessionData, key)
	}
	log.Printf("%#v", self.SessionData)
	self.ctx.SetSecureCookie(self.sessionKey, string(sessionByte), age)
}

//获取一个session值,返回值为interface,需要对获取到的值做类型断言
func (self *CookieSession) Get(key string) interface{} {
	if value, ok := self.SessionData[key]; ok {
		return value
	} else {
		return nil
	}
}

//设置一个session值
func (self *CookieSession) Set(key string, value interface{}) {
	self.SessionData[key] = value
}

//删除一个session值
func (self *CookieSession) Delete(key string) {
	delete(self.SessionData, key)
}

//清理所有的session,即将存储session的cookie删除
func (self *CookieSession) Purge() {
	self.Flush(-1)
}
