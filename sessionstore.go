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
func (self *CookieSession) restore() {
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
	sessionByte, _ := json.Marshal(self.SessionData)
	log.Printf("Flush SessionData:%v", self.SessionData)
	self.SessionData = make(map[string]interface{})
	self.handler.SetSecureCookie(self.sessionKey, string(sessionByte), self.age)
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
	//CookieSession不正确的原因在这里：
	//读取之前SD是空的
	log.Printf("Before resotre %v", self.SessionData)
	//恢复读取之后：从Cookie读取到了内容
	self.restore()
	//这里给SD赋值出现问题，如果我在Handler中连续Set两次session的话
	//第一次的Set会被忽略，因为restore的时候还是从Cookie读取，而不是当前的SD
	//详细在examples里的sessiontest.go
	log.Printf("Restored %v", self.SessionData)
	self.SessionData[key] = value
	log.Printf("Set %v", self.SessionData)
	//self.Flush()
}

//删除一个session值
func (self *CookieSession) Delete(key string) {
	self.restore()
	delete(self.SessionData, key)
	self.age = 0
	self.Flush()
}

//清理所有的session,即将存储session的cookie删除
func (self *CookieSession) Purge() {
	self.age = -1
	self.Flush()
}
