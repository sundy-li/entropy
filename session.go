package entropy

//session存储接口，实现此接口即可供框架调用
type ISessionStore interface {
	Restore(sessionId string)
	Get(key string) interface{}
	Set(key string, value interface{})
	Delete(key string)
	Purge()
	Flush()
}

//session mixin
type Session struct {
	SessionId string
	store     ISessionStore
}

func (self *Session) Set(key string, value interface{}) {
	self.store.Set(key, value)
}

func (self *Session) Get(key string) interface{} {
	return self.store.Get(key)
}

func (self *Session) Delete(key string) {
	self.store.Delete(key)
}

func (self *Session) Purge() {
	self.store.Purge()
}

func (self *Session) Restore(sessiongId string) {
	self.store.Restore(sessiongId)
}

func (self *Session) Flush() {
	self.store.Flush()
}
