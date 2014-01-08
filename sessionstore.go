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
	age         int
}

//恢复cookie中的数据到SessionData中
func (self *CookieSession) Restore() {
	sessionStr, err := self.handler.GetSecureCookie(self.sessionKey)
	if err != nil {
		//如果SessionData中有数据，就不初始化啦！！！！！
		//这个逻辑害死哥了，排了一天BUG
		//新的问题出现了，没有写入cookie，但是存在于SessionData中
		//如何重置SessionData，何时重置
		if len(self.SessionData) == 0 {
			self.SessionData = make(map[string]interface{})
		}
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
func (self *CookieSession) Flush() {
	//if len(self.SessionData) != 0 {
	log.Printf("CookieSession Flush : session store line 60 %v", self.SessionData)
	sessionByte, _ := json.Marshal(self.SessionData)
	self.SessionData = make(map[string]interface{})

	self.handler.SetSecureCookie(self.sessionKey, string(sessionByte), self.age)
	//}
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
	self.age = 0
}

//清理所有的session,即将存储session的cookie删除
func (self *CookieSession) Purge() {
	self.age = -1
}
