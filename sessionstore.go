package entropy

import (
	"encoding/json"
)

var (
	//全局session存储
	globalSessionStore ISessionStore
)

//CookieSession构造函数,使用全局单例,因为一个应用不可能要求session处于多种存储状态
func NewCookieSession(sessionKey string, handler *Handler) ISessionStore {
	if globalSessionStore == nil {
		globalSessionStore = &CookieSession{
			SessionData: make(map[string]interface{}),
			sessionKey:  sessionKey,
			handler:     handler,
		}

	}
	return globalSessionStore
}

//CookieSession 结构体
type CookieSession struct {
	SessionData map[string]interface{}
	sessionKey  string
	handler     *Handler
}

//恢复cookie中的数据到SessionData中
func (self *CookieSession) restore() {
	sessionStr, err := self.handler.GetSecureCookie(self.sessionKey)
	if err != nil {
		self.SessionData = make(map[string]interface{})
		return
	}
	err = json.Unmarshal([]byte(sessionStr), &self.SessionData)
	if err != nil {
		panic(err)
		self.SessionData = make(map[string]interface{})
		return
	}
}

//将SessionData中的数据写入到cookie中
func (self *CookieSession) flush(age int) {
	sessionByte, _ := json.Marshal(self.SessionData)
	self.handler.SetSecureCookie(self.sessionKey, string(sessionByte), age)
}

//获取一个session值,返回值为interface,需要对获取到的值做类型断言
func (self *CookieSession) Get(key string) interface{} {
	self.restore()
	if value, ok := self.SessionData[key]; ok {
		return value
	} else {
		return nil
	}
}

//设置一个session值
func (self *CookieSession) Set(key string, value interface{}) {
	self.restore()
	self.SessionData[key] = value
	self.flush(0)
}

//删除一个session值
func (self *CookieSession) Delete(key string) {
	self.restore()
	delete(self.SessionData, key)
	self.flush(0)
}

//清理所有的session,即将存储session的cookie删除
func (self *CookieSession) Purge() {
	self.flush(-1)
}
