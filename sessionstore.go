package entropy

import (
	"encoding/json"
)

var (
	globalSessionStore ISessionStore
)

func NewCookieSession(sessionKey string, handler *RequestHandler) ISessionStore {
	if globalSessionStore == nil {
		globalSessionStore = &CookieSession{
			SessionData: make(map[string]interface{}),
			sessionKey:  sessionKey,
			handler:     handler,
		}
	}
	return globalSessionStore
}

type CookieSession struct {
	SessionData map[string]interface{}
	sessionKey  string
	handler     *RequestHandler
}

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

func (self *CookieSession) flush(age int) {
	sessionByte, _ := json.Marshal(self.SessionData)
	self.handler.SetSecureCookie(self.sessionKey, string(sessionByte), age)
}

func (self *CookieSession) Get(key string) interface{} {
	self.restore()
	if value, ok := self.SessionData[key]; ok {
		return value
	} else {
		return nil
	}
}

func (self *CookieSession) Set(key string, value interface{}) {
	self.restore()
	self.SessionData[key] = value
	self.flush(0)
}

func (self *CookieSession) Delete(key string) {
	self.restore()
	delete(self.SessionData, key)
	self.flush(0)
}

func (self *CookieSession) Purge() {
	self.flush(-1)
}
